# Grimoire

**Project Grimoire** is a lightweight, CLI-first Persistence Engine and Mini-Database. It serves as the "Nervous System" for a text-based RPG, bridging the gap between **In-Memory Data Structures** (Python Dictionaries) and **Permanent File Systems** (JSON).

This project represents **Phase 0** of a multi-year polyglot roadmap, focusing on mastering Pythonic logic before refactoring into high-performance systems languages like **Golang** and **Zig**.

**File Structure:**

- `main.py` (The script)
- `grimoire.json` (The database file - created automatically by the script)

---

## 🛠 Core Features (v0.1)

- **Persistence Layer:** State is automatically serialized to and deserialized from `grimoire.json`.
- **Atomic Updates (`set`):** Direct-access modification of entity attributes (e.g., HP, Level, Class).
- **Collection Management (`add`):** Intelligent list appending for inventory and skill management.
- **Key Retrieval (`get/status`):** Precise data fetching and full character sheet visualization.
- **Cleanup (`delete`):** Manual entity or attribute removal with safe error handling.
- **Zero-Dependency:** Built strictly using the Python 3.12+ Standard Library (`sys`, `json`, `os`).

---

## 🎮 Usage Guide

Grimoire operates via positional CLI arguments. No external wrappers like `Click` or `Typer` are used.

### 1. View Character Sheet

View the current state of the world and the player. If no file exists, Grimoire initializes using a `DEFAULT_STATE`.

```bash
python main.py status
```

### 2. Update Stats (`set`)

Change a single value (Category $\rightarrow$ Key $\rightarrow$ Value). Numbers are automatically converted from strings to integers.

```bash
# Update player health
python main.py set player hp 85

# Update player name
python main.py set player name "The Traveller"
```

### 3. Manage Inventory (`add`)

Append an item to a list. This command includes a safety check to ensure the target key is a valid list.

```bash
python main.py add player inventory "rusty_dagger"
```

### 4. Search & Destroy (`get`/`delete`)

Retrieve specific values or remove data points.

```bash
# Fetch a specific value
python main.py get player hp

# Delete an attribute or entity
python main.py delete player temp_debuff
```

---

## 🏗 System Architecture: The "Bouncer Pattern"

To ensure system reliability, Grimoire follows a strict execution flow:

1. **Initialize:** Check for `grimoire.json`. If missing, load the `DEFAULT_STATE` template.
2. **The Bouncer:** Validate `sys.argv` length and existence before accessing indices to prevent `IndexError`.
3. **Type Conversion:** Check inputs with `.isdigit()` to maintain data integrity between strings and integers.
4. **Direct Access:** Update values using Dictionary Keys directly (O(1) complexity) rather than inefficient loops.
5. **Flush:** Immediately write changes back to disk with `indent=4` for human-readability.

---

## 🗺 The Multi-Year Polyglot Roadmap

The transition from Python to Systems languages is a shift in philosophy, not just syntax.

| Feature            | Python (Current)  | Go/Zig (Future)           | The Systems Lesson                                   |
| :----------------- | :---------------- | :------------------------ | :--------------------------------------------------- |
| **Data Types**     | `dict` (Dynamic)  | `struct` / `map` (Static) | Moving from "Bags of Keys" to fixed memory schemas.  |
| **File Safety**    | `with open(...)`  | `defer f.Close()`         | Shifting from "Magic" to manual resource management. |
| **Error Handling** | `try/except`      | `if err != nil`           | Treating errors as values rather than interrupts.    |
| **JSON**           | `json.load()`     | `Unmarshal` + Struct Tags | Explicitly mapping data to memory locations.         |
| **Distribution**   | Needs Interpreter | Native Binary (`.exe`)    | The power of zero-dependency compilation.            |

---

## 📈 Evolutionary Path

- [x] **v0.1: Persistence Engine** (Python Standard Lib) - _Completed_
- [x] **v0.2: Gameplay Loop** (Random Encounters, Exploration, Combat Logic)
- [x] **v0.3: Terminal HUD** (Using `os.system('clear')` and formatted ASCII)
- [ ] **v1.0: The Go Refactor** (Porting to Golang to learn Structs and Concurrency)
- [ ] **v2.0: The Binary Shift** (Refactoring the storage layer into **Zig** for raw performance)

| Feature          | Python Approach (Current) | Go Approach (Future)                      | **The Lesson**                                                                                                                                          |
| :--------------- | :------------------------ | :---------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Data Types**   | `dict = {}` (Dynamic)     | `map[string]string` (Static)              | In Python, you can put _anything_ in the DB. In Go, you will be forced to define exactly what the DB holds (Strict Typing).                             |
| **File Safety**  | `with open(...)`          | `f, err := os.Open()` + `defer f.Close()` | You will lose Python's "Context Manager" magic and learn how to manually manage memory resources using `defer`.                                         |
| **Errors**       | `try: ... except:`        | `if err != nil { return err }`            | You will stop "catching" errors and start "handling" them as values. This is the biggest philosophy shift in Go.                                        |
| **JSON**         | `json.load()` (Magic)     | `json.Unmarshal` + Struct Tags            | Python matches JSON fields automatically. In Go, you will have to manually map JSON keys to Struct fields using "tags" (e.g., `` `json:"key_name"` ``). |
| **Distribution** | Requires Python installed | Compiles to a `.exe` / Binary             | You will see the power of Go here: you can send your `.exe` to a friend, and they can run Grimoire without installing anything.                         |

---

## 🧠 Development Philosophy

This project was developed using the **Socratic Method**. No code was copy-pasted; every line was manually implemented following architectural directives to ensure a fundamental understanding of how data moves from a user's keyboard to the machine's disk.

**Project Status:** Active / Phase 0 Milestone Complete.
