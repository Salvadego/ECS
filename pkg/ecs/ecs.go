package ecs

import (
	"reflect"
	"sync"
)

// ComponentID represents a unique identifier for a component type.
type ComponentID uint64

// Components can be anything
type Component any

// EntityID represents a unique identifier for an entity.
type EntityID uint64

// Systems should implement a Update function with the delta time, and modify their components...
type System interface {
	Update(dt float64)
}

// BitSet represents a dynamic bitset for component composition.
type BitSet []ComponentID

// Set sets the bit at the given index.
func (b *BitSet) Set(index ComponentID) {
	word, bit := index/64, (index % 64)
	for len(*b) <= int(word) {
		*b = append(*b, 0)
	}
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

// ComponentTypeRegistry manages the mapping between component types and their IDs.
type ComponentTypeRegistry struct {
	mu       sync.Mutex
	typeToID map[reflect.Type]ComponentID
	nextID   ComponentID
}

func NewComponentTypeRegistry() *ComponentTypeRegistry {
	return &ComponentTypeRegistry{
		typeToID: make(map[reflect.Type]ComponentID),
	}
}

func (r *ComponentTypeRegistry) GetComponentID(t reflect.Type) ComponentID {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id, exists := r.typeToID[t]; exists {
		return id
	}
	id := r.nextID
	r.typeToID[t] = id
	r.nextID++
	return id
}

// Archetype represents a group of entities with the same component composition.
type Archetype struct {
	signature  BitSet
	entities   []EntityID
	components map[ComponentID][]Component
}

// World represents the ECS world containing all entities and archetypes.
type World struct {
	mu              sync.Mutex
	registry        *ComponentTypeRegistry
	archetypes      []*Archetype
	entityArchetype map[EntityID]*Archetype
	entityIndex     map[EntityID]int
	nextEntityID    EntityID
	systems         []System
}

func NewWorld() *World {
	return &World{
		registry:        NewComponentTypeRegistry(),
		entityArchetype: make(map[EntityID]*Archetype),
		entityIndex:     make(map[EntityID]int),
		systems:         make([]System, 0),
	}
}

// CreateEntity creates a new entity with the given components.
func (w *World) CreateEntity(components ...Component) EntityID {
	w.mu.Lock()
	defer w.mu.Unlock()

	signature := BitSet{}
	componentData := make(map[ComponentID]Component)
	for _, comp := range components {
		t := reflect.TypeOf(comp)
		id := w.registry.GetComponentID(t)
		signature.Set(id)
		componentData[id] = comp
	}

	var archetype *Archetype
	for _, arch := range w.archetypes {
		if arch.signature.Equals(signature) {
			archetype = arch
			break
		}
	}
	if archetype == nil {
		archetype = &Archetype{
			signature:  signature,
			components: make(map[ComponentID][]Component),
		}
		w.archetypes = append(w.archetypes, archetype)
	}

	entityID := w.nextEntityID
	w.nextEntityID++

	index := len(archetype.entities)
	archetype.entities = append(archetype.entities, entityID)
	for id, comp := range componentData {
		archetype.components[id] = append(archetype.components[id], comp)
	}

	w.entityArchetype[entityID] = archetype
	w.entityIndex[entityID] = index

	return entityID
}

func (w *World) AddSystems(systems ...System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, system := range systems {
		w.systems = append(w.systems, system)
	}
}

func (w *World) Update(dt float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, system := range w.systems {
		system.Update(dt)
	}
}

// GetComponent retrieves a pointer to the component of the given type for the specified entity.
func GetComponent[T any](w *World, entity EntityID) *T {
	w.mu.Lock()
	defer w.mu.Unlock()

	archetype, exists := w.entityArchetype[entity]
	if !exists {
		return nil
	}

	id := w.registry.GetComponentID(reflect.TypeOf((*T)(nil)).Elem())
	index := w.entityIndex[entity]
	comps, exists := archetype.components[id]
	if !exists || index >= len(comps) {
		return nil
	}

	comp, ok := comps[index].(T)
	if !ok {
		return nil
	}
	return &comp
}

type Filter struct {
	include []reflect.Type
	exclude []reflect.Type
}

func NewFilter() Filter {
	return Filter{}
}

func With[T any](f Filter) Filter {
	t := reflect.TypeOf((*T)(nil)).Elem()
	f.include = append(f.include, t)
	return f
}

func Without[T any](f Filter) Filter {
	t := reflect.TypeOf((*T)(nil)).Elem()
	f.exclude = append(f.exclude, t)
	return f
}

//go:generate go run gen_queries.go -template=queries.tmpl -out=queries_gen.go -max=5
