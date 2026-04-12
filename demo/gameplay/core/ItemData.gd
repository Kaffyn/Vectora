extends Resource
## Placeholder: Old ItemData class
## Migrate to new Item.gd in inventory/

class_name ItemData

var id: String = ""
var display_name: String = ""
var description: String = ""
var icon: Texture2D
var quantity: int = 1
var max_stack: int = 1
var weapon_tag: String = ""
