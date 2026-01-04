# Grimoire

### Part 1: The Python Build (The MVP)
The goal is to understand how to manipulate data structures and file I/O **without** using heavy frameworks.

**File Structure:**
*   `main.py` (The script)
*   `grimoire.json` (The database file - created automatically by the script)

**Exact Features to Implement:**

1.  **The "Initialize" Logic:**
    *   When the script runs, it must check if `grimoire.json` exists.
    *   If not, create it with an empty JSON object `{}`.
    *   If yes, load the content into a Python Dictionary.

2.  **The "Set" Command:**
    *   Input: `python main.py set <key> <value>`
    *   Action: Add the key-value pair to the dictionary and **immediately write** it back to `memory.json`.
    *   *Constraint:* For now, handle strings only.

3.  **The "Get" Command:**
    *   Input: `python main.py get <key>`
    *   Action: Look for the key. If found, print the value. If not found, print `"Error: Key not found in Grimoire."`

4.  **The "Delete" Command:**
    *   Input: `python main.py delete <key>`
    *   Action: Remove the key and save the file. Handle the crash if the key doesn't exist (try/except).

5.  **The "Help" Command (Default):**
    *   Input: `python main.py` (no args)
    *   Action: Print a list of available commands.

---

### Part 2: The Go Refactor Strategy (The Roadmap)
When you switch to Golang, you won't just copy-paste; you will be forced to change how you think. Here are the specific parts you will refactor and **why**:

| Feature | Python Approach (Current) | Go Approach (Future) | **The Lesson** |
| :--- | :--- | :--- | :--- |
| **Data Types** | `dict = {}` (Dynamic) | `map[string]string` (Static) | In Python, you can put *anything* in the DB. In Go, you will be forced to define exactly what the DB holds (Strict Typing). |
| **File Safety** | `with open(...)` | `f, err := os.Open()` + `defer f.Close()` | You will lose Python's "Context Manager" magic and learn how to manually manage memory resources using `defer`. |
| **Errors** | `try: ... except:` | `if err != nil { return err }` | You will stop "catching" errors and start "handling" them as values. This is the biggest philosophy shift in Go. |
| **JSON** | `json.load()` (Magic) | `json.Unmarshal` + Struct Tags | Python matches JSON fields automatically. In Go, you will have to manually map JSON keys to Struct fields using "tags" (e.g., `` `json:"key_name"` ``). |
| **Distribution** | Requires Python installed | Compiles to a `.exe` / Binary | You will see the power of Go here: you can send your `.exe` to a friend, and they can run Grimoire without installing anything. |
