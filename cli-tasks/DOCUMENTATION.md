# DOCUMENTATION.md

## 1. Package & Imports

`package main`

>Declares this file as the main package, which makes it an executable program.

```
import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)
```

>Standard libraries:
- `encoding/json` → marshal/unmarshal tasks to/from JSON.
- `fmt` → formatted I/O (printing messages).
- `io/ioutil` → file reading/writing.
- `os` → OS interaction (exit codes, file info, home directory).
- `path/filepath` → filename manipulation.
- `sort` → sorting slices (for priority/date).
- `strconv` → convert strings ↔ integers (task IDs).
- `strings` → join command-line arguments.
- `time` → date parsing/formatting and animations.

>Third-party:
- `github.com/spf13/cobra` → CLI framework (subcommands, flags, help).
- `github.com/spf13/viper` → config file management.

## 2. Global Variables & Cobra Root

`var cfgFile string`

>Holds the path to an optional user-specified config file.

`var rootCmd = &cobra.Command{`

Defines the root command (taskcli).

## 3. add Subcommand

```
var addCmd = &cobra.Command{
    Use:   "add [flags] <task description>",
    Short: "Add a new task",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        created, _ := cmd.Flags().GetString("date")
        due,     _ := cmd.Flags().GetString("due")
        priority,_ := cmd.Flags().GetString("priority")
        title     := strings.Join(args, " ")
        addTask(title, created, due, priority)
    },
}
```


> Use: "add [flags] <task description>"
> Flags:
> --date, -d → creation date override.
> --due,  -u → optional due date.
> --priority, -p → "low", "med", or "high".
> Run:

1. Reads flags.
2. Joins remaining args into a single title string.
3. Calls addTask(...) to persist.

## 4. edit Subcommand
```
var editCmd = &cobra.Command{
    Use:   "edit <task ID> [flags]",
    Short: "Edit a task's title, date, due date, or priority",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        id, err := strconv.Atoi(args[0])
        if err != nil {
            fmt.Fprintln(os.Stderr, "Invalid ID:", args[0])
            os.Exit(1)
        }
        title,      _ := cmd.Flags().GetString("title")
        created,    _ := cmd.Flags().GetString("date")
        newDue,     _ := cmd.Flags().GetString("due")
        newPriority,_ := cmd.Flags().GetString("priority")
        if title=="" && created=="" && newDue=="" && newPriority=="" {
            fmt.Fprintln(os.Stderr, "Nothing to edit; provide --title, --date, --due, or --priority.")
            cmd.Help()
            os.Exit(1)
        }
        editTask(id, title, created, newDue, newPriority)
    },
}
```

> Use: "edit <task ID> [flags]"
> Flags:
> --title,   -t → new task title.
> --date,    -d → new creation date.
> --due,     -u → new due date.
> --priority,-p → new priority.
> Parses the single positional arg as an integer ID, then calls editTask(...).

## 5. list Subcommand
```
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List tasks",
    Run: func(cmd *cobra.Command, args []string) {
        dateFilter, _ := cmd.Flags().GetString("date")
        sortBy,     _ := cmd.Flags().GetString("sort")
        listTasks(dateFilter, sortBy)
    },
}
```

> Use: "list"
> Flags:
> --date, -d → filter by creation date (YYYY-MM-DD or "all").
> --sort, -s → "date" or "priority" sorting.
> Calls listTasks(...) which handles filtering, sorting, and printing.

## 6. Initialization (init & initConfig)
```
func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.taskcli.yaml)")

    // Register subcommands:
    rootCmd.AddCommand(addCmd)
    rootCmd.AddCommand(editCmd)
    rootCmd.AddCommand(listCmd)
    // TODO: wire up start, done, del, clear

    // Register flags for each subcommand:
    addCmd.Flags().StringP("date",     "d", time.Now().Format("2006-01-02"), "creation date")
    addCmd.Flags().StringP("due",      "u", "", "due date")
    addCmd.Flags().StringP("priority", "p", "", "priority low|med|high")

    editCmd.Flags().StringP("title",    "t", "", "new title")
    editCmd.Flags().StringP("date",     "d", "", "new creation date")
    editCmd.Flags().StringP("due",      "u", "", "new due date")
    editCmd.Flags().StringP("priority", "p", "", "new priority")

    listCmd.Flags().StringP("date", "d", time.Now().Format("2006-01-02"), "date filter")
    listCmd.Flags().StringP("sort", "s", "", "sort by date|priority")
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        viper.AddConfigPath(home)
        viper.SetConfigName(".taskcli")
        viper.SetConfigType("yaml")
    }
    if err := viper.ReadInConfig(); err == nil {
        fmt.Println("Using config file:", viper.ConfigFileUsed())
    }
}
```

> init() runs before main():
> Sets up Cobra to call initConfig.
> Registers global --config flag and subcommands.
> Adds per-command flags.
> initConfig() locates & reads a YAML config (e.g. ~/.taskcli.yaml) if present.

## 7. Date Validation

```go
// isValidDate checks if a string is in YYYY-MM-DD format.
func isValidDate(s string) bool {
    _, err := time.Parse("2006-01-02", s)
    return err == nil
}
```

> Utility used earlier in the CSV version; still available if needed for flag validation.

## 8. Task Model & Persistence
```
const dataFile = "tasks.json"

type Task struct {
    ID         int    `json:"id"`
    Title      string `json:"title"`
    Done       bool   `json:"done"`
    InProgress bool   `json:"in_progress"`
    Created    string `json:"created"`            // YYYY-MM-DD
    Due        string `json:"due,omitempty"`      // optional due date
    Priority   string `json:"priority,omitempty"` // "low","med","high"
}
```

> Task holds all metadata.
> Stored as a JSON array in tasks.json.

```go
func loadTasks() ([]Task, error) { … }
```

> Reads the JSON file (creates an empty list if missing), unmarshals into []Task.

```go
func saveTasks(tasks []Task) error { … }
```

> Marshals to pretty JSON, writes via temporary file + rename (atomic).

func nextID(tasks []Task) int { … }

> Finds the max existing ID and returns max+1 for uniqueness.

## 9. Core Operations

### 9.1. Adding a Task

```go
func addTask(title, created, due, priority string) {
    tasks, _ := loadTasks()
    t := Task{
        ID:         nextID(tasks),
        Title:      title,
        Done:       false,
        InProgress: false,
        Created:    created,
        Due:        due,
        Priority:   priority,
    }
    tasks = append(tasks, t)
    saveTasks(tasks)
    fmt.Printf("Added task %d: %s\n", t.ID, t.Title)
}
```

> Builds a Task struct, appends it, saves, and confirms.

### 9.2. Listing Tasks

```go
func listTasks(dateFilter, sortBy string) {
    tasks, _ := loadTasks()

    // 1) Sort if requested:
    if sortBy == "priority" { … }
    else if sortBy == "date" { … }

    // 2) Filter by creation date:
    filtered := []Task{}
    for _, t := range tasks {
        if dateFilter == "all" || t.Created == dateFilter {
            filtered = append(filtered, t)
        }
    }

    // 3) Print, with status & due-date suffix:
    for _, t := range filtered {
        status := " "
        if t.Done       { status = "x" }
        else if t.InProgress { status = ">" }

        dueSuffix := ""
        if t.Due != "" {
            // overdue, due today, or tomorrow
        }
        fmt.Printf("[%s] %d: %s%s\n", status, t.ID, t.Title, dueSuffix)
    }
}
```

> Status symbol:
> " " pending
> ">" in-progress
> "x" done

> Due suffix: shows (overdue), (due today), (due tomorrow).

### 9.3. Marking Done & Animation

```go
func completeTask(id int) {
    tasks, _ := loadTasks()
    for i, t := range tasks {
        if t.ID == id {
            if !t.Done {
                tasks[i].Done = true
                fmt.Printf("Marked task %d done.\n", id)
                animateCelebrate()
            }
            break
        }
    }
    saveTasks(tasks)
}
```

> Toggles the Done flag and plays a short confetti animation:

```go
func animateCelebrate() {
    frames := []string{"🎉","✨","🎊","✨"}
    for i := 0; i < 8; i++ {
        fmt.Printf("\r%s Completed! %s", frames[i%len(frames)], frames[(i+1)%len(frames)])
        time.Sleep(100 * time.Millisecond)
    }
    fmt.Println()
}
```

### 9.4. In-Progress, Deletion & Clearing

```go
func startTask(id int) { … }   // sets InProgress=true
func deleteTask(id int) { … }  // removes a task by ID
func clearTasks() { … }        // wipes all tasks
```

> All follow the same load-mutate-save pattern and report success/failure.

### 9.5. Editing Tasks

```go
func editTask(id int, newTitle, newCreated, newDue, newPriority string) {
    tasks, _ := loadTasks()
    for i, t := range tasks {
        if t.ID == id {
            if newTitle    != "" { tasks[i].Title    = newTitle }
            if newCreated  != "" { tasks[i].Created  = newCreated }
            if newDue      != "" { tasks[i].Due      = newDue }
            if newPriority != "" { tasks[i].Priority = newPriority }
            fmt.Printf("Task %d updated.\n", id)
            break
        }
    }
    saveTasks(tasks)
}
```

> Allows partial updates—only flags provided are changed.

### 10. Help & Entry Point

```go
func usage() { … }
```

> A fallback usage printer for legacy code (Cobra handles help automatically now).

```go
func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

> Runs Cobra’s command dispatcher.
> Any command (add, list, edit, etc.) is parsed here, and its Run function invoked.

### 11. Summary

With these pieces you have:

• Data modeling with JSON persistence.

• Rich CLI via Cobra: subcommands, flags, config files.

• Core operations: add, list (filter & sort), start, done (with animation), edit, delete, clear.

• Task metadata: creation date, due date (with overdue highlighting), priority, and in-progress state.