package ecs

import (
	"math/bits"
	"reflect"
	"sync"

	"github.com/Salvadego/ECS/pkg/utils"
)

type (
	Entity      uint32
	ComponentID uint32
	Component   any
	System      interface{ Update(dt float64) }
	BitSet      uint32
	Filter      struct{ bits []BitSet }

	World struct {
		mu              sync.RWMutex
		nextEntity      Entity
		nextComponentID ComponentID

		entities    utils.Set[Entity]
		entityMasks map[Entity]Filter

		componentIDs     map[reflect.Type]ComponentID
		componentStorage map[ComponentID]map[Entity]Component

		systems []System
	}
)

func NewWorld() *World {
	return &World{
		mu:               sync.RWMutex{},
		entities:         make(utils.Set[Entity], 0),
		entityMasks:      make(map[Entity]Filter, 0),
		componentIDs:     make(map[reflect.Type]ComponentID, 0),
		componentStorage: make(map[ComponentID]map[Entity]Component, 0),
		systems:          make([]System, 0),
	}
}

func (w *World) getComponentID(componentType reflect.Type) ComponentID {
	id, exists := w.componentIDs[componentType]
	if !exists {
		id = w.nextComponentID
		w.nextComponentID++
		w.componentIDs[componentType] = id
		w.componentStorage[id] = make(map[Entity]Component)
	}
	return id
}

func (w *World) NewEntity() Entity {
	w.mu.Lock()
	defer w.mu.Unlock()

	id := w.nextEntity
	w.nextEntity++
	w.entities.Add(id)
	w.entityMasks[id] = NewFilter()
	return id
}

func (w *World) RemoveEntity(e Entity) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entities.Remove(e)

	mask := w.entityMasks[e]
	delete(w.entityMasks, e)

	for _, compID := range w.componentIDs {
		if mask.Contains(Filter{bits: []BitSet{1 << compID}}) {
			delete(w.componentStorage[compID], e)
		}
	}
}

func (w *World) AddComponent(e Entity, component Component) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.entities[e]; !exists {
		return
	}

	componentType := reflect.TypeOf(component)
	componentID := w.getComponentID(componentType)

	w.componentStorage[componentID][e] = component

	mask := w.entityMasks[e]
	mask.Add(componentID)
	w.entityMasks[e] = mask
}

func UpdateComponent[T Component](w *World, e Entity, updater func(*T)) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	var zero T
	componentType := reflect.TypeOf(zero)
	componentID, exists := w.componentIDs[componentType]
	if !exists {
		return false
	}

	component, exists := w.componentStorage[componentID][e]
	if !exists {
		return false
	}

	typed := component.(T)
	updater(&typed)
	w.componentStorage[componentID][e] = typed
	return true
}

func GetComponent[T Component](w *World, e Entity) (T, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var zero T

	if _, exists := w.entities[e]; !exists {
		return zero, false
	}

	componentType := reflect.TypeOf(zero)
	componentID, exists := w.componentIDs[componentType]
	if !exists {
		return zero, false
	}

	component, exists := w.componentStorage[componentID][e]
	if !exists {
		return zero, false
	}

	return component.(T), true
}

func RemoveComponent[T Component](w *World, e Entity) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.entities[e]; !exists {
		return
	}

	var zero T
	componentType := reflect.TypeOf(zero)
	componentID, exists := w.componentIDs[componentType]
	if !exists {
		return
	}

	delete(w.componentStorage[componentID], e)

	mask := w.entityMasks[e]
	newMask := NewFilter()
	for _, compID := range w.componentIDs {
		if compID != componentID && mask.Contains(Filter{bits: []BitSet{1 << compID}}) {
			newMask.Add(compID)
		}
	}
	w.entityMasks[e] = newMask
}

func (w *World) CreateFilter(components ...Component) Filter {
	w.mu.RLock()
	defer w.mu.RUnlock()

	filter := NewFilter()

	for _, component := range components {
		t1 := reflect.TypeOf(component)
		if id, exists := w.componentIDs[t1]; exists {
			filter.Add(id)
		}
	}

	return filter
}

func (w *World) Query(filter Filter) []Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()

	results := make([]Entity, 0)

	for entity, mask := range w.entityMasks {
		if mask.Contains(filter) {
			results = append(results, entity)
		}
	}

	return results
}

func (w *World) QueryEach(filter Filter, fn func(Entity)) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for entity, mask := range w.entityMasks {
		if mask.Contains(filter) {
			fn(entity)
		}
	}
}

func (w *World) AddSystems(systems ...System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, system := range systems {
		w.systems = append(w.systems, system)
	}
}

func (w *World) Update(dt float64) {
	for _, system := range w.systems {
		system.Update(dt)
	}
}

func NewFilter() Filter {
	return Filter{
		bits: make([]BitSet, 4),
	}
}

func (f *Filter) Add(id ComponentID) {
	index, bit := id/32, id%32
	if int(index) >= len(f.bits) {
		newBits := make([]BitSet, index+1)
		copy(newBits, f.bits)
		f.bits = newBits
	}
	f.bits[index] |= 1 << bit
}

func (f *Filter) Contains(other Filter) bool {
	for i := 0; i < len(f.bits) && i < len(other.bits); i++ {
		if other.bits[i] & ^f.bits[i] != 0 {
			return false
		}
	}
	return true
}

func (f *Filter) Matches(other Filter) bool {
	maxLen := max(len(other.bits), len(f.bits))

	for i := range maxLen {
		var fbits, obits BitSet
		if i < len(f.bits) {
			fbits = f.bits[i]
		}
		if i < len(other.bits) {
			obits = other.bits[i]
		}
		if fbits != obits {
			return false
		}
	}
	return true
}

func (f *Filter) Clear() {
	for i := range f.bits {
		f.bits[i] = 0
	}
}

func (f *Filter) Count() int {
	count := 0
	for _, v := range f.bits {
		count += bits.OnesCount32(uint32(v))
	}
	return count
}
