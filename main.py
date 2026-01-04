#!/usr/bin/env python3
"""
Grimoire - CLI RPG Engine (interactive REPL)

This file is a refreshed version with:
- Polished HUD visuals with optional ANSI color support (obeys NO_COLOR env)
- Two-column inventory display
- XP / level-up handling (simple linear curve: next_xp = level * 100)
- Admin commands operate on the current in-memory state and save it
- Small UX improvements and clearer messages
- Preserves previous gameplay semantics (explore, hunt, rest, use, save, reset, etc.)
"""

import copy
import getpass
import json
import os
import random
import shlex
import sys
import time

# === Configuration ===
SAVE_FILE = "grimoire.json"
TMP_SAVE = SAVE_FILE + ".tmp"
REST_HP_PER_SP = 25  # HP restored per 1 SP used in rest
DEFAULT_MAX_HP = 100
HUNT_BASE_SP = 1
HUNT_EXTRA_SP_MAX = 5
RESTOCK_COMMAND_INTERVAL = 20  # placeholder for future shop/restock

# === Default state ===
DEFAULT_STATE = {
    "player": {
        "name": "Traveller",
        "class": "Adventurer",
        "gold": 50,
        "hp": DEFAULT_MAX_HP,
        "max_hp": DEFAULT_MAX_HP,
        "sp": 10,
        "level": 1,
        "xp": 0,
        # Inventory is a mapping item_id -> count for stacking
        "inventory": {"torch": 1, "rusty_dagger": 1},
    },
    "meta": {
        "location": "Starting Village",
        "quests_completed": 0,
        "command_count": 0,
    },
}

# === Item catalog ===
ITEM_CATALOG = {
    "healing_potion": {
        "name": "Healing Potion",
        "hp_min": 10,
        "hp_max": 25,
        "sp_min": 1,
        "sp_max": 3,
    },
    "torch": {"name": "Torch"},
    "rusty_dagger": {"name": "Rusty Dagger"},
    # sample items referenced by enemies
    "bone_shield": {"name": "Bone Shield"},
    "ancient_coin": {"name": "Ancient Coin"},
    "coin_pouch": {"name": "Coin Pouch"},
    "wolf_pelt": {"name": "Wolf Pelt"},
    "meat": {"name": "Meat"},
    "bear_claw": {"name": "Bear Claw"},
    "orcish_blade": {"name": "Orcish Blade"},
}

# === Enemy templates ===
ENEMY_TEMPLATES = {
    "goblin": {
        "name": "Goblin",
        "hp": 8,
        "attack_min": 1,
        "attack_max": 3,
        "xp": 5,
        "gold": 3,
        "loot": [("rusty_dagger", 0.20), ("healing_potion", 0.10)],
    },
    "skeleton": {
        "name": "Skeleton",
        "hp": 10,
        "attack_min": 2,
        "attack_max": 4,
        "xp": 8,
        "gold": 5,
        "loot": [("bone_shield", 0.10), ("ancient_coin", 0.25)],
    },
    "bandit": {
        "name": "Bandit",
        "hp": 12,
        "attack_min": 2,
        "attack_max": 5,
        "xp": 10,
        "gold": 8,
        "loot": [("coin_pouch", 0.30), ("healing_potion", 0.15)],
    },
    "wolf": {
        "name": "Wolf",
        "hp": 14,
        "attack_min": 3,
        "attack_max": 6,
        "xp": 12,
        "gold": 6,
        "loot": [("wolf_pelt", 0.30), ("meat", 0.40)],
    },
    "bear": {
        "name": "Bear",
        "hp": 20,
        "attack_min": 4,
        "attack_max": 8,
        "xp": 20,
        "gold": 10,
        "loot": [("bear_claw", 0.25)],
    },
    "orc": {
        "name": "Orc",
        "hp": 25,
        "attack_min": 5,
        "attack_max": 10,
        "xp": 25,
        "gold": 15,
        "loot": [("orcish_blade", 0.15), ("coin_pouch", 0.25)],
    },
}

# === Utilities: Colors & symbols ===
NO_COLOR = os.environ.get("NO_COLOR") not in (
    None,
    "",
)  # if set to anything, disable color


# Provide simple color palette; if NO_COLOR, functions become no-ops
class _Colors:
    RED = "\x1b[31m"
    GREEN = "\x1b[32m"
    YELLOW = "\x1b[33m"
    BLUE = "\x1b[34m"
    MAGENTA = "\x1b[35m"
    CYAN = "\x1b[36m"
    BOLD = "\x1b[1m"
    RESET = "\x1b[0m"
    DIM = "\x1b[2m"


def colorize(text, color_code):
    if NO_COLOR:
        return text
    return f"{color_code}{text}{_Colors.RESET}"


HEART = "♥"
SPARK = "⚡"
STAR = "✦"


# === Persistence helpers ===
def load_game():
    """Load JSON save or return deep copy of DEFAULT_STATE if missing/corrupt."""
    if os.path.exists(SAVE_FILE):
        try:
            with open(SAVE_FILE, "r") as f:
                data = json.load(f)
            # Ensure inventory is a dict (legacy list support)
            p = data.get("player", {})
            inv = p.get("inventory")
            if isinstance(inv, list):
                # convert list to stacked dict
                stacked = {}
                for item in inv:
                    stacked[item] = stacked.get(item, 0) + 1
                data["player"]["inventory"] = stacked
            # Sanitize basic numeric fields
            try:
                data["player"]["hp"] = int(data["player"].get("hp", DEFAULT_MAX_HP))
                data["player"]["max_hp"] = int(
                    data["player"].get("max_hp", DEFAULT_MAX_HP)
                )
                data["player"]["sp"] = int(data["player"].get("sp", 0))
                data["player"]["level"] = int(data["player"].get("level", 1))
                data["player"]["xp"] = int(data["player"].get("xp", 0))
                data["player"]["gold"] = int(data["player"].get("gold", 0))
            except Exception:
                # ignore and rely on defaults where needed
                pass
            return data
        except (json.JSONDecodeError, ValueError):
            ts = int(time.time())
            badname = f"{SAVE_FILE}.corrupt.{ts}"
            try:
                os.replace(SAVE_FILE, badname)
                print(f"Warning: corrupted save moved to {badname}")
            except Exception:
                print("Warning: failed to move corrupted save; initializing defaults.")
            return copy.deepcopy(DEFAULT_STATE)
        except Exception as e:
            print(f"Warning: failed to read save file ({e}); initializing defaults.")
            return copy.deepcopy(DEFAULT_STATE)
    return copy.deepcopy(DEFAULT_STATE)


def save_game_atomic(state):
    """Write to a temp file then atomically replace the save."""
    tmp = TMP_SAVE
    try:
        with open(tmp, "w") as f:
            json.dump(state, f, indent=4)
        os.replace(tmp, SAVE_FILE)
    except Exception as e:
        # Attempt fallback
        try:
            with open(SAVE_FILE, "w") as f:
                json.dump(state, f, indent=4)
        except Exception as e2:
            print(f"Error saving game: {e2}")
        else:
            print(f"Warning: atomic save failed ({e}); used non-atomic save instead.")


# === Inventory helpers ===
def inv_get_count(state, item_id):
    return int(state["player"].get("inventory", {}).get(item_id, 0))


def inv_add(state, item_id, qty=1):
    inv = state["player"].setdefault("inventory", {})
    inv[item_id] = inv.get(item_id, 0) + int(qty)


def inv_remove(state, item_id, qty=1):
    inv = state["player"].setdefault("inventory", {})
    have = inv.get(item_id, 0)
    qty = int(qty)
    if have <= qty:
        inv.pop(item_id, None)
    else:
        inv[item_id] = have - qty


# === XP / Level handling ===
def xp_to_next(level):
    """Simple linear curve: next_xp = level * 100"""
    return int(level * 100)


def grant_xp(state, amount):
    """Add XP and handle level-ups. Returns list of level-up messages."""
    msgs = []
    p = state["player"]
    p["xp"] = int(p.get("xp", 0)) + int(amount)
    while True:
        level = int(p.get("level", 1))
        need = xp_to_next(level)
        if p["xp"] >= need:
            p["xp"] -= need
            p["level"] = level + 1
            # Grant modest benefits per level
            p["max_hp"] = int(p.get("max_hp", DEFAULT_MAX_HP)) + 10
            # Heal some on level up
            p["hp"] = min(p.get("max_hp", p["hp"]), p.get("hp", 0) + 10)
            msgs.append(
                f"You leveled up to level {p['level']}! Max HP increased to {p['max_hp']}."
            )
        else:
            break
    return msgs


# === Combat & enemy selection ===
def resolve_combat(player, enemy_template):
    """
    Combat loop:
    - Player attacks first.
    - Player damage scales with level.
    - Enemy attacks with its attack range.
    Returns: dict with outcome, player_hp (remaining), xp, gold, loot list.
    """
    player_hp = int(player.get("hp", 0))
    level = int(player.get("level", 1))
    enemy_hp = int(enemy_template.get("hp", 1))
    loot_found = []

    while player_hp > 0 and enemy_hp > 0:
        # player attack
        pmin = 1 + level
        pmax = 2 + level
        dmg = random.randint(pmin, pmax)
        enemy_hp -= dmg
        print(
            colorize(
                f"You hit the {enemy_template['name']} for {dmg} damage! (enemy {max(0, enemy_hp)} HP)",
                _Colors.GREEN,
            )
        )

        if enemy_hp <= 0:
            xp = int(enemy_template.get("xp", 0))
            gold = int(enemy_template.get("gold", 0))
            for item, chance in enemy_template.get("loot", []):
                if random.random() < chance:
                    loot_found.append(item)
            return {
                "outcome": "win",
                "player_hp": max(0, player_hp),
                "xp": xp,
                "gold": gold,
                "loot": loot_found,
            }

        # enemy attack
        emin = int(enemy_template.get("attack_min", 1))
        emax = int(enemy_template.get("attack_max", 1))
        edmg = random.randint(emin, emax)
        player_hp = max(0, player_hp - edmg)
        print(
            colorize(
                f"The {enemy_template['name']} hits you for {edmg} damage! (you {player_hp} HP)",
                _Colors.RED,
            )
        )

        if player_hp <= 0:
            return {"outcome": "lose", "player_hp": 0, "xp": 0, "gold": 0, "loot": []}

    return {
        "outcome": "lose",
        "player_hp": max(0, player_hp),
        "xp": 0,
        "gold": 0,
        "loot": [],
    }


def choose_enemy_for_location(game_state, extra_sp=0):
    """
    Choose an enemy with bias: extra_sp shifts probability towards tougher enemies.
    """
    pool = ["goblin", "skeleton", "bandit", "wolf", "bear", "orc"]
    weights = [25, 20, 15, 10, 5, 2]
    player_level = int(game_state["player"].get("level", 1))
    bias = min(max(0, int(extra_sp)), HUNT_EXTRA_SP_MAX)

    if player_level >= 3:
        # slightly shift to tougher enemies
        weights = [max(5, w - 5) for w in weights]
        weights[2] += 5  # slightly favor bandit

    if bias > 0:
        # move weight from easiest to hardest as bias increases
        weights[0] = max(0, weights[0] - bias * 8)
        weights[-1] = min(100, weights[-1] + bias * 8)

    total = sum(weights)
    roll = random.randint(1, total)
    cum = 0
    for name, w in zip(pool, weights):
        cum += w
        if roll <= cum:
            return name
    return pool[0]


# === Admin helpers ===
def require_admin_interactive(provided_password=None):
    """
    Return True if admin auth succeeds; False otherwise.
    - Uses GRIMOIRE_ADMIN_KEY env var as the secret.
    - If GRIMOIRE_ADMIN_NONINTERACTIVE=1 and GRIMOIRE_ADMIN_KEY is present,
      then if provided_password equals the key, allow non-interactive usage.
    """
    expected = os.environ.get("GRIMOIRE_ADMIN_KEY")
    if not expected:
        print(
            colorize(
                "[admin] Admin access disabled. Set GRIMOIRE_ADMIN_KEY env var to enable.",
                _Colors.YELLOW,
            )
        )
        return False

    noninteractive = os.environ.get("GRIMOIRE_ADMIN_NONINTERACTIVE") == "1"
    if noninteractive and provided_password is not None:
        if provided_password == expected:
            return True
        print(colorize("[admin] Authentication failed (non-interactive).", _Colors.RED))
        return False

    try:
        provided = getpass.getpass("[admin] Enter admin password: ")
    except Exception:
        provided = None
    if provided != expected:
        print(colorize("[admin] Authentication failed.", _Colors.RED))
        return False
    return True


def admin_set_value(state, path, value):
    """
    Set a value in the current game state by dot-separated path.
    Example: admin set player.hp 50
    Returns True if modified, False otherwise.
    """
    parts = path.split(".")
    cur = state
    for p in parts[:-1]:
        if p not in cur or not isinstance(cur[p], dict):
            print(colorize(f"[admin] Path invalid at '{p}'", _Colors.RED))
            return False
        cur = cur[p]
    key = parts[-1]
    if key not in cur:
        print(colorize(f"[admin] Key '{key}' not found in target object", _Colors.RED))
        return False
    # try to interpret numbers
    if isinstance(cur[key], int):
        try:
            cur[key] = int(value)
        except Exception:
            print(colorize(f"[admin] Value must be an integer for {path}", _Colors.RED))
            return False
    else:
        # strings or other types - set as string
        cur[key] = value
    return True


def admin_add_to_inventory(state, path, item_id, qty=1):
    """
    Add item(s) to inventory. Path should be 'player.inventory' or similar.
    """
    parts = path.split(".")
    cur = state
    for p in parts:
        if p not in cur:
            print(colorize(f"[admin] Path invalid at '{p}'", _Colors.RED))
            return False
        cur = cur[p]
    if not isinstance(cur, dict):
        print(colorize("[admin] Target is not an inventory dict", _Colors.RED))
        return False
    qty = int(qty)
    cur[item_id] = cur.get(item_id, 0) + qty
    return True


# === Command handlers ===
def cmd_status(state, args):
    """Print HUD with borders, HP/SP bars, and stacked inventory in two columns."""
    p = state["player"]
    name = p.get("name", "Player")
    pclass = p.get("class", "Adventurer")
    level = int(p.get("level", 1))
    hp = int(p.get("hp", 0))
    max_hp = int(p.get("max_hp", DEFAULT_MAX_HP))
    sp = int(p.get("sp", 0))
    gold = int(p.get("gold", 0))
    xp = int(p.get("xp", 0))
    location = state.get("meta", {}).get("location", "Unknown")

    # HUD drawing - adapt to terminal-ish width
    width = 64
    hr = "+" + "-" * (width - 2) + "+"
    title = f" {name} ({pclass}) - Lv {level} "
    right = f"{location}"
    header = f"|{title.ljust(width - 2 - len(right))}{right}|"

    print(colorize(hr, _Colors.CYAN))
    print(colorize(header, _Colors.CYAN))

    # HP bar
    bar_width = 30
    filled_hp = int((hp / max_hp) * bar_width) if max_hp > 0 else 0
    hp_bar_filled = "█" * filled_hp
    hp_bar_empty = " " * (bar_width - filled_hp)
    hp_bar = "[" + hp_bar_filled + hp_bar_empty + "]"
    hp_text = f"{HEART} {hp}/{max_hp}"
    hp_line = f"| HP {hp_bar} {hp_text.ljust(14)}|"
    print(
        colorize(
            hp_line,
            _Colors.RED
            if hp < max_hp * 0.4
            else (_Colors.YELLOW if hp < max_hp * 0.75 else _Colors.GREEN),
        )
    )

    # SP bar
    sp_width = 12
    filled_sp = min(sp, sp_width)
    sp_bar = "[" + ("●" * filled_sp) + (" " * (sp_width - filled_sp)) + "]"
    sp_text = f"{SPARK} {sp}"
    sp_line = f"| SP {sp_bar} {sp_text.ljust(18)}|"
    print(colorize(sp_line, _Colors.MAGENTA))

    # XP progress
    need = xp_to_next(level)
    pct = int((xp / need) * 100) if need > 0 else 0
    xp_bar_width = 30
    filled_xp = int((xp / need) * xp_bar_width) if need > 0 else 0
    xp_bar = "[" + "#" * filled_xp + "-" * (xp_bar_width - filled_xp) + "]"
    xp_line = f"| XP {xp_bar} {xp}/{need} ({pct}%)"
    print(colorize(xp_line.ljust(width - 1) + "|", _Colors.BLUE))

    # Resources
    res_line = (
        f"| Gold: {gold} ".ljust(30)
        + f"| Commands: {state.get('meta', {}).get('command_count', 0)}".rjust(30)
        + "|"
    )
    print(colorize(res_line, _Colors.CYAN))
    print(colorize("|" + " " * (width - 2) + "|", _Colors.CYAN))

    # Inventory header
    print(colorize("| Inventory:".ljust(width - 1) + "|", _Colors.CYAN))
    inv = state["player"].get("inventory", {}) or {}
    if not inv:
        print(colorize("|  (empty)".ljust(width - 1) + "|", _Colors.DIM))
    else:
        # Prepare two-column layout
        items = sorted(
            inv.items(), key=lambda x: (-x[1], x[0])
        )  # sort by count desc, then name
        col_width = (width - 6) // 2
        for i in range(0, len(items), 2):
            left = items[i]
            left_name = ITEM_CATALOG.get(left[0], {}).get("name", left[0])
            left_text = f"  - {left_name} x{left[1]}"
            right_text = ""
            if i + 1 < len(items):
                right = items[i + 1]
                right_name = ITEM_CATALOG.get(right[0], {}).get("name", right[0])
                right_text = f"  - {right_name} x{right[1]}"
                line = left_text.ljust(col_width) + "  " + right_text.ljust(col_width)
            else:
                line = left_text
            print(colorize(f"|{line.ljust(width - 2)}|", _Colors.DIM))
    print(colorize(hr, _Colors.CYAN))
    return False


def cmd_save(state, args):
    save_game_atomic(state)
    print(colorize("Game saved.", _Colors.GREEN))
    return False


def cmd_exit(state, args):
    save_game_atomic(state)
    print(colorize("Game saved. Exiting.", _Colors.GREEN))
    sys.exit(0)


def cmd_reset(state, args):
    confirm = input("Type 'yes' to confirm reset (this will overwrite your save): ")
    if confirm.strip().lower() == "yes":
        save_game_atomic(copy.deepcopy(DEFAULT_STATE))
        print(colorize("Game reset to defaults.", _Colors.YELLOW))
        return True
    print("Reset cancelled.")
    return False


def cmd_rest(state, args):
    """Rest: spend SP to gain HP. Usage: rest [amount]."""
    amt = 1
    if args:
        try:
            amt = int(args[0])
        except ValueError:
            print("Invalid SP amount.")
            return False
    if amt <= 0:
        print("Enter a positive SP amount to rest.")
        return False
    if state["player"].get("sp", 0) < amt:
        print("Not enough SP.")
        return False
    state["player"]["sp"] -= amt
    hp_gain = amt * REST_HP_PER_SP
    state["player"]["hp"] = min(
        state["player"].get("max_hp", DEFAULT_MAX_HP),
        state["player"].get("hp", 0) + hp_gain,
    )
    print(
        colorize(
            f"You rested using {amt} SP and recovered {hp_gain} HP.", _Colors.GREEN
        )
    )
    return True


def cmd_use(state, args):
    """Use an item by id (e.g. use healing_potion)."""
    if not args:
        print("Usage: use <item_id>")
        return False
    item_id = args[0]
    count = inv_get_count(state, item_id)
    if count <= 0:
        print(f"You don't have any {item_id} to use.")
        return False
    meta = ITEM_CATALOG.get(item_id)
    if not meta:
        print(f"Cannot use {item_id} (no effect defined).")
        return False
    # Implement healing potion
    if item_id == "healing_potion":
        hp_gain = random.randint(meta["hp_min"], meta["hp_max"])
        sp_gain = random.randint(meta["sp_min"], meta["sp_max"])
        state["player"]["hp"] = min(
            state["player"].get("max_hp", DEFAULT_MAX_HP),
            state["player"].get("hp", 0) + hp_gain,
        )
        state["player"]["sp"] = state["player"].get("sp", 0) + sp_gain
        inv_remove(state, item_id, 1)
        print(
            colorize(
                f"You used a {meta['name']}: +{hp_gain} HP, +{sp_gain} SP.",
                _Colors.GREEN,
            )
        )
        return True
    print(f"No use-effect implemented for {item_id}.")
    return False


def cmd_admin(state, args):
    """Admin operations on the current loaded state (requires GRIMOIRE_ADMIN_KEY).
    Usage:
      admin set <path> <value>                 # path example: player.hp or meta.location
      admin add inventory <path> <item> [qty]  # e.g. admin add inventory player.inventory healing_potion 2
    """
    if not args:
        print("Usage: admin <set|add> ...")
        return False

    # allow optional password as last arg for non-interactive scripts (not echoed)
    provided_password = None
    # If the caller provided --pw=xxx in args, capture it (convenience)
    for a in args[:]:
        if a.startswith("--pw="):
            provided_password = a.split("=", 1)[1]
            args.remove(a)
            break

    if not require_admin_interactive(provided_password):
        return False

    sub = args[0]
    mutated = False
    if sub == "set":
        if len(args) < 3:
            print("Usage: admin set <path> <value>")
            return False
        path = args[1]
        value = args[2]
        if admin_set_value(state, path, value):
            save_game_atomic(state)
            print(colorize(f"[admin] Set {path} = {value}", _Colors.GREEN))
            mutated = True
        else:
            return False
    elif sub == "add":
        # support: admin add player.inventory healing_potion [qty]
        if len(args) < 3:
            print("Usage: admin add <path> <item_id> [qty]")
            return False
        path = args[1]
        item_id = args[2]
        qty = int(args[3]) if len(args) >= 4 else 1
        if admin_add_to_inventory(state, path, item_id, qty):
            save_game_atomic(state)
            print(colorize(f"[admin] Added {item_id} x{qty} to {path}", _Colors.GREEN))
            mutated = True
        else:
            return False
    else:
        print("[admin] Unknown admin subcommand.")
        return False
    return mutated


def cmd_explore(state, args):
    """Explore command with random events."""
    if state["player"].get("hp", 0) <= 0:
        print("You are down (HP 0). Rest or revive first.")
        return False
    state["meta"]["command_count"] = state["meta"].get("command_count", 0) + 1
    roll = random.randint(1, 100)
    # treasure very rare
    if roll <= 2:
        print(
            colorize(
                "You found a hidden treasure chest filled with gold and items!",
                _Colors.YELLOW,
            )
        )
        state["player"]["gold"] += random.randint(100, 500)
        found_item = random.choice(["healing_potion", "rusty_dagger", "torch"])
        inv_add(state, found_item)
        print(colorize(f"You obtained a {found_item}!", _Colors.GREEN))
        return True
    # special item
    if roll <= 10:
        found_item = random.choice(["healing_potion", "torch"])
        inv_add(state, found_item)
        print(colorize(f"You explored and found a {found_item}!", _Colors.GREEN))
        return True
    # gold
    if roll <= 30:
        found_gold = random.randint(5, 50)
        state["player"]["gold"] += found_gold
        print(colorize(f"You explored and found {found_gold} gold!", _Colors.GREEN))
        return True
    # enemy
    if roll <= 50:
        extra_sp = 0
        enemy_name = choose_enemy_for_location(state, extra_sp=extra_sp)
        template = ENEMY_TEMPLATES.get(enemy_name)
        if not template:
            print("You sensed an enemy but couldn't identify it; nothing happens.")
            return False
        print(
            colorize(
                f"You encountered a {template['name']} while exploring!", _Colors.YELLOW
            )
        )
        result = resolve_combat(state["player"], template)
        state["player"]["hp"] = max(0, int(result["player_hp"]))
        if result["outcome"] == "win":
            # grant xp and gold
            xp_msgs = grant_xp(state, int(result["xp"]))
            state["player"]["gold"] += int(result["gold"])
            for it in result["loot"]:
                inv_add(state, it)
            print(
                colorize(
                    f"You defeated the {template['name']} and gained {result['xp']} XP and {result['gold']} gold!",
                    _Colors.GREEN,
                )
            )
            for m in xp_msgs:
                print(colorize(f"  {m}", _Colors.MAGENTA))
            if result["loot"]:
                print(
                    colorize(f"Looted items: {', '.join(result['loot'])}", _Colors.CYAN)
                )
            else:
                print(colorize("No loot found.", _Colors.DIM))
        else:
            print(colorize("You were defeated and lost the fight.", _Colors.RED))
        return True
    # nothing
    print("You explored but found nothing of interest.")
    return False


def cmd_hunt(state, args):
    """
    Hunt:
      - baseline costs HUNT_BASE_SP (1) SP
      - optional extra_sp (int) stakes for better loot & tougher enemies
      - if victory and player's hp ended at 0, revive to 1 (as per rule)
    Usage: hunt [extra_sp]
    """
    if state["player"].get("hp", 0) <= 0:
        print("You are down (HP 0). Rest or revive first.")
        return False

    extra_sp = 0
    if args:
        try:
            extra_sp = int(args[0])
        except ValueError:
            print("Invalid extra SP amount.")
            return False
    extra_sp = max(0, min(HUNT_EXTRA_SP_MAX, extra_sp))

    total_sp_cost = HUNT_BASE_SP + extra_sp
    if state["player"].get("sp", 0) < total_sp_cost:
        print("Not enough SP to hunt with that stake.")
        return False

    # Deduct SP
    state["player"]["sp"] -= total_sp_cost
    state["meta"]["command_count"] = state["meta"].get("command_count", 0) + 1

    # Determine enemy and scale rewards
    enemy_name = choose_enemy_for_location(state, extra_sp=extra_sp)
    template = ENEMY_TEMPLATES.get(enemy_name)
    if not template:
        print("No enemy encountered.")
        return False

    print(
        colorize(f"You encountered a {template['name']} while hunting!", _Colors.YELLOW)
    )
    # Simulate combat - but hunt stakes increase reward multiplier
    result = resolve_combat(state["player"], template)

    # Apply results
    state["player"]["hp"] = max(0, int(result["player_hp"]))
    if result["outcome"] == "win":
        # scale rewards by extra_sp: each extra SP increases rewards by 25% per SP
        multiplier = 1.0 + 0.25 * extra_sp
        gained_xp = int(result["xp"] * multiplier)
        gained_gold = int(result["gold"] * multiplier)
        xp_msgs = grant_xp(state, gained_xp)
        state["player"]["gold"] += gained_gold
        for it in result["loot"]:
            inv_add(state, it)
        # revive-to-1 rule
        if state["player"]["hp"] == 0:
            state["player"]["hp"] = 1
            print(
                colorize(
                    "You survived the hunt and were revived to 1 HP.", _Colors.YELLOW
                )
            )
        print(
            colorize(
                f"You defeated the {template['name']} and gained {gained_xp} XP and {gained_gold} gold!",
                _Colors.GREEN,
            )
        )
        for m in xp_msgs:
            print(colorize(f"  {m}", _Colors.MAGENTA))
        if result["loot"]:
            print(colorize(f"Looted items: {', '.join(result['loot'])}", _Colors.CYAN))
    else:
        print(colorize("You were defeated and lost the fight.", _Colors.RED))
    return True


# === REPL and CLI ===
def print_usage():
    print("Commands:")
    print("  play                Enter interactive mode")
    print("  status              Show HUD status")
    print("  explore             Explore the world")
    print("  hunt [extra_sp]     Hunt (costs SP; extra_sp increases rewards)")
    print("  use <item_id>       Use an item from inventory (e.g. healing_potion)")
    print(
        f"  rest [sp]           Rest and convert SP into HP (each SP -> {REST_HP_PER_SP} HP)"
    )
    print("  save                Save game")
    print("  reset               Reset to defaults (confirmation required)")
    print("  exit / quit         Save and exit")
    print("  admin ...           Admin commands (requires GRIMOIRE_ADMIN_KEY env var)")


def repl_loop():
    state = load_game()
    handlers = {
        "status": cmd_status,
        "explore": cmd_explore,
        "hunt": cmd_hunt,
        "save": cmd_save,
        "exit": cmd_exit,
        "quit": cmd_exit,
        "reset": cmd_reset,
        "rest": cmd_rest,
        "use": cmd_use,
        "admin": cmd_admin,
    }

    print("Entering interactive play mode. Type 'help' for commands.")
    while True:
        try:
            line = input("> ")
        except (EOFError, KeyboardInterrupt):
            print("\nExiting and saving...")
            save_game_atomic(state)
            break
        if not line:
            continue
        parts = shlex.split(line)
        if not parts:
            continue
        cmd = parts[0].lower()
        args = parts[1:]
        if cmd in ("help", "?"):
            print_usage()
            continue
        handler = handlers.get(cmd)
        if not handler:
            print("Unknown command. Type 'help' for a list of commands.")
            continue
        mutated = handler(state, args)
        if mutated:
            save_game_atomic(state)


def main():
    # If no args or 'play', enter REPL
    if len(sys.argv) == 1 or (len(sys.argv) >= 2 and sys.argv[1] == "play"):
        repl_loop()
        return

    # One-shot commands
    command = sys.argv[1]
    state = load_game()
    if command == "status":
        cmd_status(state, sys.argv[2:])
    elif command == "explore":
        mutated = cmd_explore(state, sys.argv[2:])
        if mutated:
            save_game_atomic(state)
    elif command == "hunt":
        mutated = cmd_hunt(state, sys.argv[2:])
        if mutated:
            save_game_atomic(state)
    elif command == "save":
        cmd_save(state, sys.argv[2:])
    elif command == "reset":
        if cmd_reset(state, sys.argv[2:]):
            # reload defaults into memory
            state = load_game()
    elif command == "rest":
        if cmd_rest(state, sys.argv[2:]):
            save_game_atomic(state)
    elif command == "use":
        if cmd_use(state, sys.argv[2:]):
            save_game_atomic(state)
    elif command == "admin":
        if cmd_admin(state, sys.argv[2:]):
            # admin handlers already saved or performed actions
            pass
    else:
        print(f"Unknown command: {command}")


if __name__ == "__main__":
    main()
