extends Node
## Inventory system for managing items

class_name Inventory

const MAX_SLOTS: int = 20

var items: Array[Item] = []
var max_slots: int = MAX_SLOTS

signal item_added(item: Item, slot: int)
signal item_removed(item: Item, slot: int)
signal item_quantity_changed(item: Item, slot: int)
signal inventory_full

func _ready() -> void:
	# Initialize empty slots
	items.resize(max_slots)

func add_item(item: Item) -> bool:
	"""Add item to inventory, return true if successful"""
	if item.quantity <= 0:
		return false

	# Try to stack with existing items
	for i in range(items.size()):
		if items[i] and items[i].can_stack_with(item):
			var overflow = items[i].add_quantity(item.quantity)
			item_quantity_changed.emit(items[i], i)
			if overflow > 0:
				item.quantity = overflow
				return add_item(item)  # Recursively add overflow
			return true

	# Find empty slot
	for i in range(items.size()):
		if !items[i]:
			items[i] = item
			item_added.emit(item, i)
			return true

	# Inventory full
	inventory_full.emit()
	return false

func remove_item(slot: int) -> bool:
	"""Remove item from slot"""
	if slot < 0 or slot >= items.size() or !items[slot]:
		return false

	var item = items[slot]
	items[slot] = null
	item_removed.emit(item, slot)
	return true

func remove_item_quantity(slot: int, quantity: int) -> bool:
	"""Remove specific quantity from item"""
	if slot < 0 or slot >= items.size() or !items[slot]:
		return false

	var item = items[slot]
	if item.quantity < quantity:
		return false

	item.quantity -= quantity
	if item.quantity <= 0:
		items[slot] = null
		item_removed.emit(item, slot)
	else:
		item_quantity_changed.emit(item, slot)
	return true

func get_item(slot: int) -> Item:
	"""Get item at slot"""
	if slot >= 0 and slot < items.size():
		return items[slot]
	return null

func get_item_by_id(item_id: String) -> Item:
	"""Find first item with matching ID"""
	for item in items:
		if item and item.id == item_id:
			return item
	return null

func has_item(item_id: String) -> bool:
	"""Check if inventory contains item"""
	return get_item_by_id(item_id) != null

func get_free_slots() -> int:
	"""Count empty slots"""
	var count = 0
	for item in items:
		if !item:
			count += 1
	return count

func is_full() -> bool:
	return get_free_slots() == 0

func clear() -> void:
	"""Remove all items"""
	for i in range(items.size()):
		if items[i]:
			remove_item(i)

func get_all_items() -> Array[Item]:
	"""Get list of non-empty items"""
	var result: Array[Item] = []
	for item in items:
		if item:
			result.append(item)
	return result
