package ecs

import (
	"sync"
	"unsafe"
)

// ComponentID represents a unique identifier for a component type.
type ComponentID uint64

// Components can be anything
type Component interface {
	ID() ComponentID
}

// EntityID represents a unique identifier for an entity.
type EntityID uint64

// Systems should implement a Update function with the delta time, and modify their components...
type System interface {
	Update(dt float64)
}

// BitSet represents a dynamic bitset for component composition.
type BitSet [2]ComponentID

// Set sets the bit at the given index.
func (b *BitSet) Set(index ComponentID) {
	word, bit := index/64, (index % 64)
	(*b)[word] |= 1 << bit
}

// Has checks if the bit at the given index is set.
func (b BitSet) Has(index ComponentID) bool {
	word, bit := index/64, uint(index%64)
	if int(word) >= len(b) {
		return false
	}
	return (b[word] & (1 << bit)) != 0
}

// Equals checks if two BitSets are equal.
func (b BitSet) Equals(other BitSet) bool {
	// Fast path - direct comparison
	return b[0] == other[0] && b[1] == other[1]
}

func (b BitSet) ContainsAll(other BitSet) bool {
	return (b[0]&other[0]) == other[0] && (b[1]&other[1]) == other[1]
}

func (b BitSet) Intersects(other BitSet) bool {
	return (b[0]&other[0] != 0) || (b[1]&other[1] != 0)
}

// Hash generates a hash value for the BitSet for map lookup
func (b BitSet) Hash() uint64 {
	return uint64(b[0]) ^ (uint64(b[1]) << 32)
}

func (b BitSet) Indices() []ComponentID {
	// Pre-count bits to allocate exact size
	count := 0
	for _, word := range b {
		x := word
		for x != 0 {
			count++
			x &= x - 1
		}
	}

	ids := make([]ComponentID, 0, count)
	for wordIdx, word := range b {
		if word == 0 {
			continue
		}
		for bit := uint(0); bit < 64; bit++ {
			if (word & (1 << bit)) != 0 {
				ids = append(ids, ComponentID(wordIdx*64+int(bit)))
			}
		}
	}
	return ids
}

// ComponentTypeInfo stores type information for a component type
type ComponentTypeInfo struct {
	id       ComponentID
	size     uintptr
	typeName string
	pool     sync.Pool
}

var componentTypes = make(map[ComponentID]*ComponentTypeInfo)

// RegisterComponentType registers information about a component type
func RegisterComponentType[T Component](id ComponentID) {
	var zero T
	size := unsafe.Sizeof(zero)
	componentTypes[id] = &ComponentTypeInfo{
		id:   id,
		size: size,
		pool: sync.Pool{
			New: func() any {
				return make([]Component, 0, 64)
			},
		},
	}
}

// EntityData stores entity information
type EntityData struct {
	archetype *Archetype
	index     int
}

// ComponentSlot stores components of a single type
type ComponentSlot struct {
	id   ComponentID
	data []Component
}

// Archetype represents a group of entities with the same component composition.
type Archetype struct {
	mu          sync.RWMutex
	signature   BitSet
	entities    []EntityID
	components  []ComponentSlot
	compIndex   map[ComponentID]int
	entityIndex map[EntityID]int
}

// GetComponentData provides direct access to a component array
func (a *Archetype) GetComponentData(id ComponentID) ([]Component, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if idx, ok := a.compIndex[id]; ok && idx < len(a.components) {
		return a.components[idx].data, true
	}
	return nil, false
}

// AddEntity adds an entity to this archetype
func (a *Archetype) AddEntity(entityID EntityID, componentMap map[ComponentID]Component) int {
	a.mu.Lock()
	defer a.mu.Unlock()

	index := len(a.entities)
	a.entities = append(a.entities, entityID)
	a.entityIndex[entityID] = index

	for id, comp := range componentMap {
		if idx, ok := a.compIndex[id]; ok {
			a.components[idx].data = append(a.components[idx].data, comp)
		}
	}

	return index
}

// Query cache to avoid recreating similar queries
type queryCache struct {
	mu     sync.RWMutex
	cache  map[uint64][][]Component
	filter Filter
}

// World represents the ECS world containing all entities and archetypes.
type World struct {
	mu                    sync.RWMutex
	archetypes            []*Archetype
	archetypeMap          map[uint64]*Archetype
	archetypesByComponent map[ComponentID][]*Archetype
	entityData            map[EntityID]EntityData
	nextEntityID          EntityID
	systems               []System
	queryCache            map[uint64]*queryCache
}

// NewWorld creates a new World instance.
func NewWorld() *World {
	return &World{
		entityData:            make(map[EntityID]EntityData, 1024),
		archetypeMap:          make(map[uint64]*Archetype, 64),
		archetypesByComponent: make(map[ComponentID][]*Archetype, 32),
		systems:               make([]System, 0, 16),
		queryCache:            make(map[uint64]*queryCache),
	}
}

// registerArchetype adds a new archetype to the world and updates indexes
func (w *World) registerArchetype(archetype *Archetype) {
	w.archetypes = append(w.archetypes, archetype)
	hash := archetype.signature.Hash()
	w.archetypeMap[hash] = archetype

	for id := range archetype.compIndex {
		w.archetypesByComponent[id] = append(w.archetypesByComponent[id], archetype)
	}
}

// getOrCreateArchetype gets an existing archetype or creates a new one if it doesn't exist
func (w *World) getOrCreateArchetype(signature BitSet, componentMap map[ComponentID]Component) *Archetype {
	hash := signature.Hash()
	archetype, exists := w.archetypeMap[hash]

	if exists && archetype.signature.Equals(signature) {
		return archetype
	}

	compArray := make([]ComponentSlot, 0, len(componentMap))
	compIndex := make(map[ComponentID]int, len(componentMap))

	i := 0
	for id := range componentMap {
		compIndex[id] = i

		var data []Component
		if info, ok := componentTypes[id]; ok {
			data = info.pool.Get().([]Component)
			data = data[:0]
		} else {
			data = make([]Component, 0, 64)
		}

		compArray = append(compArray, ComponentSlot{
			id:   id,
			data: data,
		})
		i++
	}

	archetype = &Archetype{
		signature:   signature,
		components:  compArray,
		compIndex:   compIndex,
		entityIndex: make(map[EntityID]int, 64),
	}

	w.registerArchetype(archetype)
	return archetype
}

// CreateEntity creates a new entity with the given components.
func (w *World) CreateEntity(components ...Component) EntityID {
	signature := BitSet{}
	componentMap := make(map[ComponentID]Component, len(components))
	for _, comp := range components {
		id := comp.ID()
		signature.Set(id)
		componentMap[id] = comp
	}

	w.mu.Lock()
	entityID := w.nextEntityID
	w.nextEntityID++

	archetype := w.getOrCreateArchetype(signature, componentMap)

	index := archetype.AddEntity(entityID, componentMap)

	w.entityData[entityID] = EntityData{
		archetype: archetype,
		index:     index,
	}

	w.mu.Unlock()
	return entityID
}

func (w *World) AddSystems(systems ...System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if cap(w.systems)-len(w.systems) < len(systems) {
		newSystems := make([]System, len(w.systems), len(w.systems)+len(systems))
		copy(newSystems, w.systems)
		w.systems = newSystems
	}

	for _, system := range systems {
		w.systems = append(w.systems, system)
	}
}

// Update runs all systems
func (w *World) Update(dt float64) {
	w.mu.RLock()
	systems := w.systems
	w.mu.RUnlock()

	for _, system := range systems {
		system.Update(dt)
	}
}

// GetComponent retrieves a component for an entity
func GetComponent[T Component](w *World, entity EntityID) T {
	w.mu.RLock()
	data, exists := w.entityData[entity]
	w.mu.RUnlock()

	var zero T
	if !exists {
		return zero
	}

	id := zero.ID()

	data.archetype.mu.RLock()
	defer data.archetype.mu.RUnlock()

	if idx, ok := data.archetype.compIndex[id]; ok {
		components := data.archetype.components[idx].data
		if data.index < len(components) {
			return components[data.index].(T)
		}
	}

	return zero
}

type Filter struct {
	include BitSet
	exclude BitSet
}

func NewFilter(include ...ComponentID) Filter {
	var filter Filter
	for _, id := range include {
		filter.include.Set(id)
	}
	return filter
}

func (f *Filter) Without(ids ...ComponentID) *Filter {
	for _, id := range ids {
		f.exclude.Set(id)
	}
	return f
}

// QueryIterator allows for efficient iteration over query results
type QueryIterator struct {
	archetypes       []*Archetype
	includeIDs       []ComponentID
	currentArchetype int
	currentEntity    int
	componentArrays  [][]Component
	row              []Component
}

// Next advances to the next result, returns false when done
func (qi *QueryIterator) Next() bool {
	for qi.currentArchetype < len(qi.archetypes) {
		arch := qi.archetypes[qi.currentArchetype]

		if qi.componentArrays == nil {
			arch.mu.RLock()
			qi.componentArrays = make([][]Component, len(qi.includeIDs))
			allPresent := true

			for i, id := range qi.includeIDs {
				if comps, ok := arch.GetComponentData(id); ok {
					qi.componentArrays[i] = comps
				} else {
					allPresent = false
					break
				}
			}

			if !allPresent {
				arch.mu.RUnlock()
				qi.currentArchetype++
				qi.currentEntity = 0
				qi.componentArrays = nil
				continue
			}

			entityCount := len(arch.entities)
			arch.mu.RUnlock()

			if entityCount == 0 {
				qi.currentArchetype++
				qi.currentEntity = 0
				qi.componentArrays = nil
				continue
			}
		}

		if qi.currentEntity >= len(qi.componentArrays[0]) {
			qi.currentArchetype++
			qi.currentEntity = 0
			qi.componentArrays = nil
			continue
		}

		if qi.row == nil {
			qi.row = make([]Component, len(qi.includeIDs))
		}

		valid := true
		for i, comps := range qi.componentArrays {
			if qi.currentEntity < len(comps) {
				qi.row[i] = comps[qi.currentEntity]
			} else {
				valid = false
				break
			}
		}

		qi.currentEntity++

		if valid {
			return true
		}
	}

	return false
}

// Row returns the current result row
func (qi *QueryIterator) Row() []Component {
	return qi.row
}

// Iterator returns an iterator for the query results
func (f Filter) Iterator(w *World) *QueryIterator {
	w.mu.RLock()
	defer w.mu.RUnlock()

	includeIDs := f.include.Indices()
	if len(includeIDs) == 0 {
		return &QueryIterator{archetypes: []*Archetype{}}
	}

	// Find archetypes with fewest matching entities first
	var candidateArchetypes []*Archetype
	minCount := -1

	for _, id := range includeIDs {
		archetypes, exists := w.archetypesByComponent[id]
		if !exists {
			return &QueryIterator{archetypes: []*Archetype{}}
		}

		count := len(archetypes)
		if minCount == -1 || count < minCount {
			minCount = count
			candidateArchetypes = archetypes
		}
	}

	matchingArchetypes := make([]*Archetype, 0, len(candidateArchetypes))

	for _, arch := range candidateArchetypes {
		if f.includeMatch(arch.signature) && !f.excludeMatch(arch.signature) {
			matchingArchetypes = append(matchingArchetypes, arch)
		}
	}

	return &QueryIterator{
		archetypes: matchingArchetypes,
		includeIDs: includeIDs,
	}
}

// Query returns all matching component rows
func (f Filter) Query(w *World) [][]Component {
	w.mu.RLock()

	cacheKey := f.include.Hash() ^ (f.exclude.Hash() << 1)
	if cache, ok := w.queryCache[cacheKey]; ok {
		cache.mu.RLock()
		w.mu.RUnlock()
		result := cache.cache[cacheKey]
		cache.mu.RUnlock()
		if result != nil {
			return result
		}
	}
	w.mu.RUnlock()

	it := f.Iterator(w)
	result := make([][]Component, 0, 64)

	for it.Next() {
		row := make([]Component, len(it.row))
		copy(row, it.row)
		result = append(result, row)
	}

	w.mu.Lock()
	if _, ok := w.queryCache[cacheKey]; !ok {
		w.queryCache[cacheKey] = &queryCache{
			cache:  make(map[uint64][][]Component),
			filter: f,
		}
	}
	cache := w.queryCache[cacheKey]
	w.mu.Unlock()

	cache.mu.Lock()
	cache.cache[cacheKey] = result
	cache.mu.Unlock()

	return result
}

func (f Filter) includeMatch(sig BitSet) bool {
	return sig.ContainsAll(f.include)
}

func (f Filter) excludeMatch(sig BitSet) bool {
	return f.exclude.Intersects(sig)
}

func (w *World) ClearQueryCache() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.queryCache = make(map[uint64]*queryCache)
}
