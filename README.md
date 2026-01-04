# Grimoire

Project Grimoire is a small, CLI-first persistence engine and interactive RPG shell. It provides a minimal, deterministic bridge between an in-memory Python state (simple `dict` objects) and a human-editable JSON file on disk. The implementation is intentionally compact and zero-dependency so it can be studied, extended, and later ported to systems languages (Go, Zig) as part of a polyglot learning roadmap.

This README is synchronized to the current `main.py` and documents the REPL, one-shot commands, gameplay mechanics, persistence behaviour, admin features, constants, and useful examples.

---

Table of contents

- Overview
- Files
- Quickstart
- CLI: one-shot commands
- REPL: interactive play mode
- Core gameplay features
  - Player state and inventory
  - Items and use-effects
  - Combat and enemies
  - XP & leveling
  - Explore & Hunt semantics
  - Rest & SP
- Persistence & savefile handling
- Admin mode
- Environment variables & flags
- Constants and tuning knobs
- Debugging & notes
- Examples

---

Overview

Grimoire exposes both an interactive REPL (play mode) and simple one-shot commands so you can embed it into scripts or use it interactively. The in-memory structure is persisted to `grimoire.json` after mutating operations. The on-disk format is plain JSON for easy manual inspection.

Files

- `main.py` — the engine and CLI entrypoint (implements REPL, commands, persistence, combat and admin).
- `grimoire.json` — the JSON save file (created automatically if missing).
- Temporary atomic write file used during saves: `grimoire.json.tmp` (handled internally).

Quickstart

Requirements: Python 3.12+ (no external pip packages).

Show status (one-shot):

```
python main.py status
```

Enter interactive REPL:

```
python main.py play
# or simply
python main.py
```

One-shot commands can be run directly from the shell; see the "CLI" section below.

---

CLI: one-shot commands

The CLI accepts positional arguments. If you run `python main.py` with no arguments it enters the interactive REPL.

Supported one-shot commands (callable directly from the shell):

- `status` — print the HUD / character sheet
- `explore` — explore the world once (may find gold, items, or enemies)
- `hunt [extra_sp]` — perform a hunt (costs SP; optional stake `extra_sp` increases risk/reward)
- `use <item_id>` — use an item (e.g. `healing_potion`)
- `rest [sp]` — spend SP to restore HP (`REST_HP_PER_SP` HP per SP)
- `save` — force save to disk
- `reset` — reset the save to the default state (requires confirmation)
- `admin ...` — run admin operations (requires `GRIMOIRE_ADMIN_KEY` — see Admin mode)
- `exit` / `quit` — save and exit (in one-shot usage, these are primarily used in REPL)

Examples:

```
python main.py explore
python main.py hunt 2
python main.py use healing_potion
python main.py rest 2
```

---

REPL: interactive play mode

Start the REPL:

```
python main.py play
# or
python main.py
```

Inside the REPL you can type commands described above repeatedly. Use `help` or `?` to show a short usage summary. The REPL reads commands via `input`, parses them with `shlex.split`, and dispatches to handlers. Mutating operations automatically trigger saves at the end of the command loop (the handler returns a boolean indicating whether to persist).

When you exit the REPL (Ctrl-D, Ctrl-C, or `exit`), the engine saves the current state.

---

Core gameplay features

Player state and inventory

The saved top-level structure includes a `player` object and `meta` info. The default state contains:

- `player.name` — string
- `player.class` — string
- `player.gold` — int
- `player.hp`, `player.max_hp` — ints
- `player.sp` — stamina points (int)
- `player.level`, `player.xp` — ints
- `player.inventory` — stacking inventory represented as a mapping: `item_id -> count` (a dict). This allows easy stacking/consumption of items.

The default `player.inventory` in `DEFAULT_STATE` contains stacked items like:

```json
{ "torch": 1, "rusty_dagger": 1 }
```

Important: legacy saves that store `inventory` as a list are converted to a stacked dict automatically on load.

Items and use-effects

The engine contains an `ITEM_CATALOG` that defines items and their use effects where applicable. For example:

- `healing_potion` — has `hp_min`, `hp_max`, `sp_min`, `sp_max` and restores random HP/SP when used.
- `torch`, `rusty_dagger`, and other items exist as catalog entries (some are informational/loot-only).

Using an item:

```
python main.py use healing_potion
```

If the `item_id` is present and a use-effect is defined (currently `healing_potion`), the effect will be applied and the item decremented in the stacked inventory.

Inventory helpers implemented:

- `inv_get_count(state, item_id)` — returns the stacked count
- `inv_add(state, item_id, qty=1)` — increments a stacked entry
- `inv_remove(state, item_id, qty=1)` — decrements and removes when count reaches zero

Combat and enemies

Enemies are defined in `ENEMY_TEMPLATES` with fields like `hp`, `attack_min`, `attack_max`, `xp`, `gold`, and `loot` (a list of `(item_id, chance)` pairs).

Combat resolution (function `resolve_combat(player, enemy_template)`):

- Player attacks first on each round.
- Player damage scales with player `level`: player's damage range is `(1 + level)` to `(2 + level)`.
- If the enemy dies, the player wins: they gain XP and gold and possibly loot items.
- If player HP drops to 0 the player loses that combat.
- Combat prints a message per hit (uses colorized output when allowed).

Enemy selection (`choose_enemy_for_location(game_state, extra_sp)`):

- Base enemy pool: `goblin, skeleton, bandit, wolf, bear, orc` with default weights.
- Player level and `extra_sp` bias (used by `hunt`) shift weights toward tougher enemies.
- `extra_sp` is capped at `HUNT_EXTRA_SP_MAX` and is used to increase the chance of higher-tier enemies.

XP & leveling

XP curve is simple and linear:

- `xp_to_next(level)` returns `level * 100`.

Level-up behavior (`grant_xp(state, amount)`):

- Add XP to `player.xp`, and while XP exceeds the threshold:
  - Deduct threshold and increment `player.level`.
  - Increase `player.max_hp` by `+10`.
  - Heal the player by `+10` HP (bounded by `max_hp`).
- `grant_xp` returns level-up messages that are displayed when leveling occurs.

Explore & Hunt semantics

`explore`:

- A single exploration roll (random 1–100) leads to:
  - Very rare treasure (gold + item)
  - Rare item find (like a `healing_potion` or `torch`)
  - Common gold find
  - Enemy encounter (resolves combat)
  - Nothing

`hunt [extra_sp]`:

- Costs `HUNT_BASE_SP` (default 1) plus optional `extra_sp` (stakes, capped by `HUNT_EXTRA_SP_MAX`).
- Deducts SP before the hunt.
- `extra_sp` increases enemy difficulty (via selection bias) and scales rewards on victory.
- Rewards are multiplied by `1 + 0.25 * extra_sp` (25% extra per stake).
- If the player wins but their HP ended at 0, they are revived to 1 HP (a safety rule for hunts).

Rest & SP

`rest [sp]`:

- Spend SP to restore HP.
- Each SP spent restores `REST_HP_PER_SP` HP (configurable constant).
- The command rejects invalid inputs and prevents spending more SP than the player has.

---

Persistence & savefile handling

- Save file: `grimoire.json`
- Atomic save strategy: write to `grimoire.json.tmp` and then `os.replace(tmp, grimoire.json)` to do an atomic rename. A fallback non-atomic write is attempted if atomic replace fails.
- `load_game()` behavior:
  - If `grimoire.json` exists, attempts to parse it with `json.load()`.
  - If the `player.inventory` is a list (legacy), it converts to a stacked dict.
  - It sanitizes a few numeric fields (`hp`, `max_hp`, `sp`, `level`, `xp`, `gold`) by trying to cast to `int`.
  - If the file is corrupt JSON, it renames the corrupt file to `grimoire.json.corrupt.<timestamp>` and initializes defaults.
  - If the file is missing, returns a deep copy of `DEFAULT_STATE`.

Because the engine rewrites the entire file on each save, keep backups if you run bulk edits.

---

Admin mode

Admin operations allow direct manipulation of the in-memory state and are gated by an environment secret:

- Enable admin: set the environment variable `GRIMOIRE_ADMIN_KEY` to a desired password.
- Non-interactive admin (CI/scripts): set `GRIMOIRE_ADMIN_NONINTERACTIVE=1` and provide `--pw=<secret>` in the admin arguments to authenticate without interactive prompt.

Supported admin subcommands:

- `admin set <path> <value>`
  - `path` is dot-separated (e.g. `player.hp` or `meta.location`).
  - Only existing integer fields are coerced to integers; non-int fields are set as strings.
  - Requires the target key to already exist (it will not create nested structure keys).
  - Example:
    ```
    python main.py admin set player.hp 50 --pw=secret
    ```

- `admin add <path> <item_id> [qty]`
  - `path` should point to an inventory dict (e.g. `player.inventory`).
  - Adds the given `item_id` with optional quantity (default 1).
  - Example:
    ```
    python main.py admin add player.inventory healing_potion 2 --pw=secret
    ```

Admin helpers are implemented as `admin_set_value` and `admin_add_to_inventory` and they save immediately when successful.

---

Environment variables & flags

- `NO_COLOR` — if present (any non-empty value), color output is disabled.
- `GRIMOIRE_ADMIN_KEY` — required to enable admin operations (password).
- `GRIMOIRE_ADMIN_NONINTERACTIVE=1` — allow noninteractive admin if you supply `--pw=<secret>`.

---

Constants and tuning knobs

- `SAVE_FILE = "grimoire.json"` — savefile name
- `TMP_SAVE = SAVE_FILE + ".tmp"` — atomic temporary file
- `REST_HP_PER_SP = 25` — HP recovered per SP when using `rest`
- `DEFAULT_MAX_HP = 100`
- `HUNT_BASE_SP = 1`
- `HUNT_EXTRA_SP_MAX = 5`
- XP curve: `xp_to_next(level) = level * 100`
- Enemy templates and item catalog are defined in `main.py` constants: `ENEMY_TEMPLATES`, `ITEM_CATALOG`.

These are located near the top of `main.py` and are safe to tune for gameplay variations.

---

Debugging & notes

- If you receive a "corrupted save" warning, the corrupt save file will be renamed (timestamped) and defaults will be loaded; recover the corrupt data manually if needed.
- To inspect JSON quickly:
  ```
  python -c "import json;print(json.dumps(json.load(open('grimoire.json')), indent=2))"
  ```
- The engine assumes a single-writer model (no file locking). If you require concurrent access, add file locking or run a single server instance to mediate writes.
- Colorized output is purely cosmetic; disabling it with `NO_COLOR` is useful in CI or when piping output.

---

Examples

Start a REPL session:

```
python main.py
# > help
# > status
# > explore
# > hunt 2
# > use healing_potion
# > rest 1
# > save
# > exit
```

One-shot hunt with stakes:

```
python main.py hunt 3
```

Admin (non-interactive) add healing potions:

```
export GRIMOIRE_ADMIN_KEY=secret
export GRIMOIRE_ADMIN_NONINTERACTIVE=1
python main.py admin add player.inventory healing_potion 3 --pw=secret
```

Reset to defaults (one-shot):

```
python main.py reset
# prompts for confirmation: type 'yes' to perform reset
```

---

Roadmap

This project is intentionally small and experiential in nature. Next improvements that preserve the zero-dependency core:

- Add unit tests for command handlers (`status`, `explore`, `hunt`, `use`, `rest`, `admin`).
- Add file locking to support multi-process safety.
- Improve numeric parsing to accept negative numbers and floats where appropriate.
- Add a configurable HUD width and more display options (ASCII art, themes).
- Port to Go to learn static typing and compile-time checks, then optionally refactor storage layer to Zig for memory control and speed.

---

License & contact

This project is a learning exercise; no license is included by default. If you intend to reuse or distribute, add a LICENSE file to reflect your preferred terms.
