package ecs

import (
	"sync"
)

// ComponentID represents a unique identifier for a component type.
type ComponentID uint64

// Components can be anything
type Component interface {
	ID() ComponentID
}

type ComponentSlot struct {
	id   ComponentID
	data []Component
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
	maxLen := max(len(other), len(b))
	for i := range maxLen {
		var aWord, bWord ComponentID
		if i < len(b) {
			aWord = b[i]
		}
		if i < len(other) {
			bWord = other[i]
		}
		if aWord != bWord {
			return false
		}
	}
	return true
}

func (b BitSet) ContainsAll(other BitSet) bool {
	for i := range other {
		if i >= len(b) || (b[i]&other[i]) != other[i] {
			return false
		}
	}
	return true
}

func (b BitSet) Intersects(other BitSet) bool {
	for i := range b {
		if i < len(other) && (b[i]&other[i]) != 0 {
			return true
		}
	}
	return false
}

// Hash generates a hash value for the BitSet for map lookup
func (b BitSet) Hash() uint64 {
	var hash uint64
	for i, word := range b {
		hash ^= uint64(word) << (i * 8 % 56) // Use modulo to avoid overflow
	}
	return hash
}

func (b BitSet) Indices() []ComponentID {
	var ids []ComponentID
	for wordIdx, word := range b {
		if word == 0 {
			continue // Skip words with no bits set for efficiency
		}
		for bit := range 64 {
			if (word & (1 << uint(bit))) != 0 {
				ids = append(ids, ComponentID(wordIdx*64+bit))
			}
		}
	}
	return ids
}

type EntityData struct {
	archetype *Archetype
	index     int
}

// Archetype represents a group of entities with the same component composition.
type Archetype struct {
	signature BitSet
	entities  []EntityID
	slots     []ComponentSlot
	slotIndex map[ComponentID]int
}

func (a *Archetype) GetSlot(id ComponentID) ([]Component, bool) {
	if idx, ok := a.slotIndex[id]; ok && idx < len(a.slots) {
		return a.slots[idx].data, true
	}
	return nil, false
}

// World represents the ECS world containing all entities and archetypes.
type World struct {
	mu                    sync.RWMutex
	archetypes            []*Archetype
	archetypeMap          map[uint64]*Archetype        // Fast archetype lookup by signature hash
	archetypesByComponent map[ComponentID][]*Archetype // Index archetypes by component for faster queries
	entityData            map[EntityID]EntityData      // Combined entity data (replaces two maps)
	nextEntityID          EntityID
	systems               []System
}

// NewWorld creates a new World instance.
func NewWorld() *World {
	return &World{
		entityData:            make(map[EntityID]EntityData),
		archetypeMap:          make(map[uint64]*Archetype),
		archetypesByComponent: make(map[ComponentID][]*Archetype),
		systems:               make([]System, 0),
	}
}

// registerArchetype adds a new archetype to the world and updates indexes
func (w *World) registerArchetype(archetype *Archetype) {
	w.archetypes = append(w.archetypes, archetype)
	hash := archetype.signature.Hash()
	w.archetypeMap[hash] = archetype

	// Index by component for faster queries
	for _, id := range archetype.signature.Indices() {
		w.archetypesByComponent[id] = append(w.archetypesByComponent[id], archetype)
	}
}

// CreateEntity creates a new entity with the given components.
func (w *World) CreateEntity(components ...Component) EntityID {
	w.mu.Lock()
	defer w.mu.Unlock()

	signature := BitSet{}
	componentMap := make(map[ComponentID]Component)
	for _, comp := range components {
		id := comp.ID()
		signature.Set(id)
		componentMap[id] = comp
	}

	hash := signature.Hash()
	archetype, exists := w.archetypeMap[hash]

	if exists && !archetype.signature.Equals(signature) {
		exists = false
	}

	if !exists {
		slots := make([]ComponentSlot, 0, len(componentMap))
		slotIndex := make(map[ComponentID]int, len(componentMap))

		i := 0
		for id := range componentMap {
			slotIndex[id] = i
			slots = append(slots, ComponentSlot{
				id:   id,
				data: make([]Component, 0, 64),
			})
			i++
		}

		archetype = &Archetype{
			signature: signature,
			slots:     slots,
			slotIndex: slotIndex,
		}
		w.registerArchetype(archetype)
	}

	entityID := w.nextEntityID
	w.nextEntityID++

	index := len(archetype.entities)
	archetype.entities = append(archetype.entities, entityID)

	// Add components to their respective slots
	for id, comp := range componentMap {
		if slotIdx, ok := archetype.slotIndex[id]; ok && slotIdx < len(archetype.slots) {
			archetype.slots[slotIdx].data = append(archetype.slots[slotIdx].data, comp)
		}
	}

	w.entityData[entityID] = EntityData{
		archetype: archetype,
		index:     index,
	}

	return entityID
}

func (w *World) AddSystems(systems ...System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, system := range systems {
		w.systems = append(w.systems, system)
	}
}

// Update now allows systems to run in parallel if they're marked as safe for parallel execution
func (w *World) Update(dt float64) {
	w.mu.RLock()         // Use read lock to allow concurrent access for systems
	systems := w.systems // Copy to avoid mutex during iteration
	w.mu.RUnlock()

	// Sequential update for now - could be extended to parallel if systems implement a marker interface
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
	components, ok := data.archetype.GetSlot(id)
	if !ok || data.index >= len(components) {
		return zero
	}

	return components[data.index].(T)
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

// Query optimized to use the component indices
func (f Filter) Query(w *World) [][]Component {
	w.mu.RLock()
	defer w.mu.RUnlock()

	includeIDs := f.include.Indices()
	if len(includeIDs) == 0 {
		return nil
	}

	// Find archetypes with fewest entities first
	var candidateArchetypes []*Archetype
	minCount := -1

	for _, id := range includeIDs {
		archetypes, exists := w.archetypesByComponent[id]
		if !exists {
			return nil
		}

		count := len(archetypes)
		if minCount == -1 || count < minCount {
			minCount = count
			candidateArchetypes = archetypes
		}
	}

	totalEntities := 0
	matchingArchetypes := make([]*Archetype, 0, len(candidateArchetypes))

	for _, arch := range candidateArchetypes {
		if f.includeMatch(arch.signature) && !f.excludeMatch(arch.signature) {
			matchingArchetypes = append(matchingArchetypes, arch)
			totalEntities += len(arch.entities)
		}
	}

	result := make([][]Component, 0, totalEntities)

	rowBuffer := make([]Component, len(includeIDs))

	for _, arch := range matchingArchetypes {
		slots := make([][]Component, len(includeIDs))
		for i, id := range includeIDs {
			comps, ok := arch.GetSlot(id)
			if ok {
				slots[i] = comps
			}
		}

		for entityIdx := range arch.entities {
			valid := true
			for i, comps := range slots {
				if comps == nil || entityIdx >= len(comps) {
					valid = false
					break
				}
				rowBuffer[i] = comps[entityIdx]
			}

			if valid {
				row := make([]Component, len(includeIDs))
				copy(row, rowBuffer)
				result = append(result, row)
			}
		}
	}

	return result
}

func (f Filter) includeMatch(sig BitSet) bool {
	return sig.ContainsAll(f.include)
}

func (f Filter) excludeMatch(sig BitSet) bool {
	return f.exclude.Intersects(sig)
}
