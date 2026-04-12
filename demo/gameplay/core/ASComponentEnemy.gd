extends Node
## ASComponent wrapper for enemies - integrates Ability System with enemy combat
## Manages ability activation, effects, and ability-based damage scaling

class_name ASComponentEnemy

var owner_node: Node = null
var container: Node = null
var attribute_set: Node = null

# Signal compatibility
signal tag_changed(tag: StringName, added: bool)
signal ability_activated(ability: ASAbility)
signal ability_deactivated(ability: ASAbility)

func _ready() -> void:
	owner_node = get_parent()
	print("[ASComponentEnemy] Initialized for %s" % owner_node.name)

## Get ability power from attributes
func get_power() -> float:
	if attribute_set:
		return attribute_set.get_attribute_value_by_tag(&"attack_power")
	return 1.0

## Get defense from attributes
func get_defense() -> float:
	if attribute_set:
		return attribute_set.get_attribute_value_by_tag(&"defense")
	return 0.0

## Apply damage and trigger effects
func take_damage(amount: float) -> void:
	"""Handle damage taken"""
	if owner_node and owner_node.has_method("take_damage"):
		owner_node.take_damage(amount)

## Apply an effect by tag
func apply_effect(effect_tag: String) -> void:
	"""Apply ability effect by tag"""
	if container:
		var tag = StringName(effect_tag)
		# Try to activate effect ability
		if container.has_ability_with_tag(tag):
			var ability = container.get_ability_with_tag(tag)
			if ability:
				ability.activate(self)
				ability_activated.emit(ability)

## Get current tags (for compatibility)
func get_tags() -> Array:
	var tags: Array = []
	if container:
		for ability in container.get_abilities():
			if ability and ability.tags:
				tags.append_array(ability.tags)
	return tags

## Check if has tag
func has_tag(tag: StringName) -> bool:
	if container:
		return container.has_ability_with_tag(tag)
	return false

## Add tag (compatibility - actual tagging handled by abilities)
func add_tag(tag: StringName) -> void:
	pass

## Remove tag (compatibility)
func remove_tag(tag: StringName) -> void:
	pass

## Try activate ability by tag
func try_activate_ability_by_tag(tag: StringName) -> bool:
	if container and container.has_ability_with_tag(tag):
		var ability = container.get_ability_with_tag(tag)
		if ability:
			ability.activate(self)
			ability_activated.emit(ability)
			return true
	return false

## Get attribute value
func get_attribute_value_by_tag(tag: StringName) -> float:
	if attribute_set:
		return attribute_set.get_attribute_value_by_tag(tag)
	return 0.0

## Set attribute value
func set_attribute_base_value_by_tag(tag: StringName, value: float) -> void:
	if attribute_set:
		attribute_set.set_attribute_base_value_by_tag(tag, value)

## Register node (for ASDelivery integration)
func register_node(tag: StringName, node: Node) -> void:
	pass

## Apply container (for ability system)
func apply_container(new_container: ASAbilityContainer) -> void:
	container = new_container
	if container:
		print("[ASComponentEnemy] Container applied with %d abilities" % container.get_abilities().size())

## Unlock ability by resource
func unlock_ability_by_resource(ability: ASAbility) -> void:
	if container and ability:
		container.register_ability(ability)

## Is ability active
func is_ability_active(tag: StringName) -> bool:
	if container:
		var ability = container.get_ability_with_tag(tag)
		if ability:
			return ability.is_active
	return false

## Cancel ability by tag
func cancel_ability_by_tag(tag: StringName) -> void:
	if container:
		var ability = container.get_ability_with_tag(tag)
		if ability:
			ability.deactivate()
			ability_deactivated.emit(ability)

## Apply effect by tag
func apply_effect_by_tag(tag: StringName) -> void:
	apply_effect(String(tag))

## Perform ability by name
func perform_ability(ability_name: String) -> bool:
	"""Attempt to perform an ability by name"""
	if container:
		var ability = container.get_ability_by_name(ability_name)
		if ability:
			ability.activate(self)
			ability_activated.emit(ability)
			return true
	return false

## Get or create attribute set
func _setup_attributes() -> void:
	if not attribute_set:
		attribute_set = ASAttributeSet.new()
		add_child(attribute_set)
		# Initialize basic attributes
		attribute_set.set_attribute_base_value_by_tag(&"attack_power", 5.0)
		attribute_set.set_attribute_base_value_by_tag(&"defense", 0.0)
