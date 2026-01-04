import json
import os
import sys

SAVE_FILE = "grimoire.json"
DEFAULT_STATE = {
    "player": {
        "name": "Traveller",
        "class": "Adventurer",
        "hp": 100,
        "level": 1,
        "inventory": ["torch", "rusty dagger"],
    }
}


def load_game():
    if os.path.exists(SAVE_FILE):
        with open(SAVE_FILE, "r") as f:
            return json.load(f)
    return DEFAULT_STATE.copy()


def save_game(data):
    with open(SAVE_FILE, "w") as f:
        json.dump(data, f, indent=4)


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python main.py <command> [options]")
        sys.exit(1)
    command = sys.argv[1]
    # 1. STATUS
    if command == "status":
        game_state = load_game()
        print("Current game state:")
        print(json.dumps(game_state, indent=4))
    # 2. SET VALUE
    elif command == "set":
        game_state = load_game()
        if len(sys.argv) < 5:
            print("Usage: python main.py set <category> <value> <key>]")
            sys.exit(1)
        category, key, value = sys.argv[2:5]

        if value.isdigit():
            value = int(value)
        if category in game_state and key in game_state[category]:
            game_state[category][key] = value
            save_game(game_state)
            print(f"Set {key} to {value}.")
        else:
            print(f"Error: Category {category} or key {key} not found.")

    # 3. ADD ITEM TO INVENTORY
    elif command == "add":
        game_state = load_game()
        if len(sys.argv) < 5:
            print("Usage: python main.py add <category> <key> <item>")
            sys.exit(1)
        category, key, item = sys.argv[2:5]
        if category in game_state and key in game_state[category]:
            target = game_state[category][key]
            if isinstance(target, list):
                target.append(item)
                save_game(game_state)
                print(f"Added {item} to {category}.{key}.")
            else:
                print(f"Error: {key} is not a list!")
        else:
            print(f"Error: {category} or {key} not found.")

    # 4. UNKNOWN
    else:
        print(f"Unknown command: {command}")
