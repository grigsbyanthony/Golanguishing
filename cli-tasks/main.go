package main

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

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "taskcli",
    Short: "A CLI task manager",
    Long:  "TaskCLI is a simple to-do list manager with JSON persistence.",
}

var addCmd = &cobra.Command{
    Use:   "add [flags] <task description>",
    Short: "Add a new task",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        created, _ := cmd.Flags().GetString("date")
        due, _ := cmd.Flags().GetString("due")
        priority, _ := cmd.Flags().GetString("priority")
        title := strings.Join(args, " ")
        addTask(title, created, due, priority)
    },
}

var editCmd = &cobra.Command{
    Use:   "edit <task ID> [flags]",
    Short: "Edit a task's title or date",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        id, err := strconv.Atoi(args[0])
        if err != nil {
            fmt.Fprintln(os.Stderr, "Invalid ID:", args[0])
            os.Exit(1)
        }
        title, _ := cmd.Flags().GetString("title")
        created, _ := cmd.Flags().GetString("date")
        newDue, _ := cmd.Flags().GetString("due")
        newPriority, _ := cmd.Flags().GetString("priority")
        if title == "" && created == "" && newDue == "" && newPriority == "" {
            fmt.Fprintln(os.Stderr, "Nothing to edit; provide --title, --date, --due, or --priority.")
            cmd.Help()
            os.Exit(1)
        }
        editTask(id, title, created, newDue, newPriority)
    },
}

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List tasks",
    Run: func(cmd *cobra.Command, args []string) {
        dateFilter, _ := cmd.Flags().GetString("date")
        sortBy, _ := cmd.Flags().GetString("sort")
        listTasks(dateFilter, sortBy)
    },
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.taskcli.yaml)")
    // Here we add subcommands
    rootCmd.AddCommand(addCmd)
    rootCmd.AddCommand(editCmd)
    rootCmd.AddCommand(listCmd)
    // TODO: add other commands (start, done, del, clear, etc.)

    addCmd.Flags().StringP("date", "d", time.Now().Format("2006-01-02"), "creation date for the task")
    addCmd.Flags().StringP("due", "u", "", "due date for the task (YYYY-MM-DD)")
    addCmd.Flags().StringP("priority", "p", "", "priority for the task (low,med,high)")
    editCmd.Flags().StringP("title", "t", "", "new title for the task")
    editCmd.Flags().StringP("date", "d", "", "new date for the task (YYYY-MM-DD)")
    editCmd.Flags().StringP("due", "u", "", "new due date for the task (YYYY-MM-DD)")
    editCmd.Flags().StringP("priority", "p", "", "new priority for the task (low,med,high)")
    listCmd.Flags().StringP("date", "d", time.Now().Format("2006-01-02"), "date to filter tasks (YYYY-MM-DD or 'all')")
    listCmd.Flags().StringP("sort", "s", "", "sort tasks by 'date' or 'priority'")
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

// isValidDate checks if a string is in YYYY-MM-DD format.
func isValidDate(s string) bool {
    _, err := time.Parse("2006-01-02", s)
    return err == nil
}

const dataFile = "tasks.json"

type Task struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Done      bool   `json:"done"`
    InProgress bool  `json:"in_progress"`
    Created   string `json:"created"`
    Due       string `json:"due,omitempty"`
    Priority  string `json:"priority,omitempty"`
}

func loadTasks() ([]Task, error) {
    // If file doesn't exist, start with an empty list
    if _, err := os.Stat(dataFile); os.IsNotExist(err) {
        return []Task{}, nil
    }
    b, err := ioutil.ReadFile(dataFile)
    if err != nil {
        return nil, err
    }
    var tasks []Task
    if err := json.Unmarshal(b, &tasks); err != nil {
        return nil, err
    }
    return tasks, nil
}

func saveTasks(tasks []Task) error {
    b, err := json.MarshalIndent(tasks, "", "  ")
    if err != nil {
        return err
    }
    // Write atomically
    tmpFile := dataFile + ".tmp"
    if err := ioutil.WriteFile(tmpFile, b, 0644); err != nil {
        return err
    }
    return os.Rename(tmpFile, dataFile)
}

func nextID(tasks []Task) int {
    max := 0
    for _, t := range tasks {
        if t.ID > max {
            max = t.ID
        }
    }
    return max + 1
}

func addTask(title, created, due, priority string) {
    tasks, err := loadTasks()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error loading tasks:", err)
        os.Exit(1)
    }
    t := Task{
        ID:        nextID(tasks),
        Title:     title,
        Done:      false,
        InProgress: false,
        Created:   created,
        Due:       due,
        Priority:  priority,
    }
    tasks = append(tasks, t)
    if err := saveTasks(tasks); err != nil {
        fmt.Fprintln(os.Stderr, "Error saving tasks:", err)
        os.Exit(1)
    }
    fmt.Printf("Added task %d: %s\n", t.ID, t.Title)
}

func listTasks(dateFilter, sortBy string) {
    tasks, err := loadTasks()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error loading tasks:", err)
        os.Exit(1)
    }

    // sort tasks if requested
    if sortBy == "priority" {
        order := map[string]int{"high": 0, "med": 1, "medium": 1, "low": 2, "": 3}
        sort.Slice(tasks, func(i, j int) bool {
            return order[tasks[i].Priority] < order[tasks[j].Priority]
        })
    } else if sortBy == "date" {
        sort.Slice(tasks, func(i, j int) bool {
            return tasks[i].Created < tasks[j].Created
        })
    }

    filtered := make([]Task, 0)
    for _, t := range tasks {
        if dateFilter == "all" || t.Created == dateFilter {
            filtered = append(filtered, t)
        }
    }
    tasks = filtered
    if len(tasks) == 0 {
        if dateFilter == "all" {
            fmt.Println("No tasks found.")
        } else {
            fmt.Printf("No tasks found for %s.\n", dateFilter)
        }
        return
    }
    for _, t := range tasks {
        status := " "
        if t.Done {
            status = "x"
        } else if t.InProgress {
            status = ">"
        }
        // Determine due status suffix
        dueSuffix := ""
        if t.Due != "" {
            if dueTime, err := time.Parse("2006-01-02", t.Due); err == nil {
                today := time.Now().Truncate(24 * time.Hour)
                diff := int(dueTime.Sub(today).Hours() / 24)
                switch {
                case diff < 0:
                    dueSuffix = " (overdue)"
                case diff == 0:
                    dueSuffix = " (due today)"
                case diff == 1:
                    dueSuffix = " (due tomorrow)"
                }
            }
        }
        fmt.Printf("[%s] %d: %s%s\n", status, t.ID, t.Title, dueSuffix)
    }
}

func completeTask(id int) {
    tasks, err := loadTasks()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error loading tasks:", err)
        os.Exit(1)
    }
    found := false
    for i, t := range tasks {
        if t.ID == id {
            if t.Done {
                fmt.Printf("Task %d is already done.\n", id)
            } else {
                tasks[i].Done = true
                fmt.Printf("Marked task %d done.\n", id)
                animateCelebrate()
            }
            found = true
            break
        }
    }
    if !found {
        fmt.Printf("No task with ID %d.\n", id)
        return
    }
    if err := saveTasks(tasks); err != nil {
        fmt.Fprintln(os.Stderr, "Error saving tasks:", err)
        os.Exit(1)
    }
}

// startTask marks a task as in-progress.
func startTask(id int) {
    tasks, err := loadTasks()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error loading tasks:", err)
        os.Exit(1)
    }
    found := false
    for i, t := range tasks {
        if t.ID == id {
            if t.Done {
                fmt.Printf("Cannot start task %d; it is already done.\n", id)
            } else if t.InProgress {
                fmt.Printf("Task %d is already in progress.\n", id)
            } else {
                tasks[i].InProgress = true
                fmt.Printf("Task %d marked as in-progress.\n", id)
            }
            found = true
            break
        }
    }
    if !found {
        fmt.Printf("No task with ID %d.\n", id)
        return
    }
    if err := saveTasks(tasks); err != nil {
        fmt.Fprintln(os.Stderr, "Error saving tasks:", err)
        os.Exit(1)
    }
}

func deleteTask(id int) {
    tasks, err := loadTasks()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error loading tasks:", err)
        os.Exit(1)
    }
    newTasks := make([]Task, 0, len(tasks))
    found := false
    for _, t := range tasks {
        if t.ID == id {
            found = true
            continue
        }
        newTasks = append(newTasks, t)
    }
    if !found {
        fmt.Printf("No task with ID %d.\n", id)
        return
    }
    if err := saveTasks(newTasks); err != nil {
        fmt.Fprintln(os.Stderr, "Error saving tasks:", err)
        os.Exit(1)
    }
    fmt.Printf("Deleted task %d.\n", id)
}

// clearTasks removes all tasks by saving an empty list.
func clearTasks() {
    if err := saveTasks([]Task{}); err != nil {
        fmt.Fprintln(os.Stderr, "Error clearing tasks:", err)
        os.Exit(1)
    }
    fmt.Println("All tasks cleared.")
}

// editTask updates a task's title and/or creation date.
func editTask(id int, newTitle, newCreated, newDue, newPriority string) {
    tasks, err := loadTasks()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error loading tasks:", err)
        os.Exit(1)
    }
    found := false
    for i, t := range tasks {
        if t.ID == id {
            if newTitle != "" {
                tasks[i].Title = newTitle
            }
            if newCreated != "" {
                tasks[i].Created = newCreated
            }
            if newDue != "" {
                tasks[i].Due = newDue
            }
            if newPriority != "" {
                tasks[i].Priority = newPriority
            }
            fmt.Printf("Task %d updated.\n", id)
            found = true
            break
        }
    }
    if !found {
        fmt.Printf("No task with ID %d.\n", id)
        return
    }
    if err := saveTasks(tasks); err != nil {
        fmt.Fprintln(os.Stderr, "Error saving tasks:", err)
        os.Exit(1)
    }
}

// animateCelebrate prints a brief confetti animation in the terminal.
func animateCelebrate() {
    frames := []string{"🎉", "✨", "🎊", "✨"}
    // play a short loop of confetti frames
    for i := 0; i < 8; i++ {
        fmt.Printf("\r%s Completed! %s", frames[i%len(frames)], frames[(i+1)%len(frames)])
        time.Sleep(100 * time.Millisecond)
    }
    fmt.Println() // move to the next line after animation
}

func usage() {
    prog := filepath.Base(os.Args[0])
    fmt.Printf(`Usage:
  %s add [YYYY-MM-DD] <task description>    Add a new task (optionally with specific date)
  %s edit <task ID> [--title new title] [--date YYYY-MM-DD]   Edit a task
  %s list [YYYY-MM-DD|all]      List tasks for a date (default today) or all
  %s start <task ID>          Mark task as in-progress
  %s done <task ID>            Mark task as done
  %s del <task ID>             Delete a task
  %s clear                      Clear all tasks
  %s help                      Show this help message
`, prog, prog, prog, prog, prog, prog, prog, prog)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}