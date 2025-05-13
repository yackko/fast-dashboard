package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const defaultProjectName = "mydashboard"

// sanitizeName converts a string to a version suitable for Go package/directory names.
// Example: "My Awesome Project" -> "my_awesome_project"
func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	// Remove any characters not suitable for package names (simplified)
	var result strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			result.WriteRune(r)
		}
	}
	if result.String() == "" {
		return defaultProjectName // Fallback if sanitization results in empty string
	}
	return result.String()
}

// titleCaseFormat formats a name for use in Go function names.
// Example: "my ideas" -> "MyIdeas"
// Example: "to-do list" -> "ToDoList"
func titleCaseFormat(name string) string {
	name = strings.ReplaceAll(name, "-", " ") // Treat hyphens as spaces
	name = strings.ReplaceAll(name, "_", " ") // Treat underscores as spaces
	words := strings.Fields(name)
	var titleCasedWords []string
	for _, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			titleCasedWords = append(titleCasedWords, string(runes))
		}
	}
	return strings.Join(titleCasedWords, "")
}

// mainGoContent generates the content for cmd/<projectName>/main.go
func mainGoContent(moduleName string, tabDetails []map[string]string) string {
	var tabItems []string
	for _, tab := range tabDetails {
		// Tab name for display (original or user-friendly)
		displayName := tab["displayName"]
		// Function name (e.g., MakeMyTasksUI)
		funcName := "Make" + tab["titleCaseName"] + "UI"
		tabItems = append(tabItems, fmt.Sprintf(`container.NewTabItem("%s", ui.%s(myWindow))`, displayName, funcName))
	}

	return fmt.Sprintf(`package main

import (
	"%s/internal/ui" // Import from our internal UI package

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Dashboard") // Window title can be dynamic too

	tabs := container.NewAppTabs(
		%s,
	)

	// You can set the window title based on the project name if desired
	// myWindow.SetTitle("%s") // Pass project name here

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(700, 500))
	myWindow.ShowAndRun()
}
`, moduleName, strings.Join(tabItems, ",\n\t\t"))
}

// genericTabUIContent generates the content for a tab's UI file (e.g., internal/ui/tasks.go)
func genericTabUIContent(tabTitleCaseName, tabVarName, tabDisplayName string) string {
	return fmt.Sprintf(`package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// %sData holds the list of items for the "%s" tab.
// For a real application, you'd load/save this data.
var %sData = []string{"Sample Item 1 for %s", "Sample Item 2 for %s"}

// Make%sUI creates and returns the canvas object for the "%s" tab.
func Make%sUI(win fyne.Window) fyne.CanvasObject {
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter new %s...")

	var itemList *widget.List

	itemList = widget.NewList(
		func() int {
			return len(%sData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template item")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(%sData[i])
		},
	)

	addButton := widget.NewButton("Add %s", func() {
		if input.Text != "" {
			%sData = append(%sData, input.Text)
			itemList.Refresh()
			input.SetText("")
		}
	})

	inputBox := container.NewHBox(input, addButton)
	return container.NewBorder(inputBox, nil, nil, nil, itemList)
}
`, tabVarName, tabDisplayName, tabVarName, tabDisplayName, tabDisplayName, tabTitleCaseName, tabDisplayName, tabTitleCaseName, strings.ToLower(tabDisplayName), tabVarName, tabVarName, strings.Title(strings.ToLower(tabDisplayName)), tabVarName, tabVarName)
}

// gitignoreContent provides a standard .gitignore for Go projects.
func gitignoreContent() string {
	return `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool
*.out

# Dependency directories (e.g., vendor)
vendor/

# Go workspace file
go.work
go.work.sum

# Environment variables file
.env

# IDE / Editor specific
.vscode/
.idea/
*.swp
*~
`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// --- Get Project Name ---
	fmt.Print("Enter the name/type for your dashboard (e.g., Life Dashboard, Project Tracker): ")
	projectNameInput, _ := reader.ReadString('\n')
	projectNameInput = strings.TrimSpace(projectNameInput)
	if projectNameInput == "" {
		projectNameInput = defaultProjectName
		fmt.Printf("No project name entered, using default: %s\n", projectNameInput)
	}
	// moduleName will be used for 'go mod init <moduleName>' and import paths
	moduleName := sanitizeName(projectNameInput) // e.g., "life_dashboard"

	// --- Get Tab Names ---
	fmt.Print("Enter the names for your initial tabs, separated by commas (e.g., Ideas, To-Do, Shopping List): ")
	tabsInput, _ := reader.ReadString('\n')
	tabsInput = strings.TrimSpace(tabsInput)

	var tabNamesRaw []string
	if tabsInput != "" {
		tabNamesRaw = strings.Split(tabsInput, ",")
	} else {
		tabNamesRaw = []string{"Items"} // Default if no tabs are entered
		fmt.Println("No tabs entered, creating a default 'Items' tab.")
	}

	var tabDetails []map[string]string // To store processed tab info

	for _, rawName := range tabNamesRaw {
		trimmedName := strings.TrimSpace(rawName)
		if trimmedName == "" {
			continue
		}
		details := map[string]string{
			"displayName":   trimmedName,                                  // "Shopping List"
			"titleCaseName": titleCaseFormat(trimmedName),                 // "ShoppingList"
			"varName":       sanitizeName(trimmedName) + "Data",           // "shopping_listData"
			"fileName":      sanitizeName(trimmedName) + ".go",            // "shopping_list.go"
		}
		tabDetails = append(tabDetails, details)
	}

	if len(tabDetails) == 0 { // Should not happen if default "Items" is used, but as a safeguard
		fmt.Println("No valid tab names provided. Exiting.")
		os.Exit(1)
	}

	fmt.Printf("\nCreating project structure for: %s (Module: %s)\n", projectNameInput, moduleName)

	// --- 1. Create root project directory (using moduleName for consistency) ---
	if err := os.Mkdir(moduleName, 0755); err != nil {
		if !os.IsExist(err) {
			fmt.Printf("Error creating root directory %s: %v\n", moduleName, err)
			os.Exit(1)
		}
		fmt.Printf("Directory %s already exists. Files might be overwritten or skipped if they exist.\n", moduleName)
	}

	// --- 2. Create cmd/<moduleName>/ directory ---
	cmdDir := filepath.Join(moduleName, "cmd", moduleName)
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		fmt.Printf("Error creating cmd directory %s: %v\n", cmdDir, err)
		os.Exit(1)
	}

	// --- 3. Create internal/ui/ directory ---
	internalUiDir := filepath.Join(moduleName, "internal", "ui")
	if err := os.MkdirAll(internalUiDir, 0755); err != nil {
		fmt.Printf("Error creating internal/ui directory %s: %v\n", internalUiDir, err)
		os.Exit(1)
	}

	// --- 4. Write cmd/<moduleName>/main.go ---
	mainGoPath := filepath.Join(cmdDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent(moduleName, tabDetails)), 0644); err != nil {
		fmt.Printf("Error writing main.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created: %s\n", mainGoPath)

	// --- 5. Write UI files for each tab ---
	for _, tab := range tabDetails {
		uiFilePath := filepath.Join(internalUiDir, tab["fileName"])
		uiContent := genericTabUIContent(tab["titleCaseName"], tab["varName"], tab["displayName"])
		if err := os.WriteFile(uiFilePath, []byte(uiContent), 0644); err != nil {
			fmt.Printf("Error writing %s: %v\n", uiFilePath, err)
			os.Exit(1) // Stop if one file fails
		}
		fmt.Printf("Created: %s\n", uiFilePath)
	}

	// --- 6. Write .gitignore ---
	gitignorePath := filepath.Join(moduleName, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent()), 0644); err != nil {
		fmt.Printf("Error writing .gitignore: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created: %s\n", gitignorePath)

	fmt.Println("\nProject structure created successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("1. Navigate to the project directory: cd %s\n", moduleName)
	fmt.Printf("2. Initialize Go modules: go mod init %s\n", moduleName)
	fmt.Println("   (If your project is hosted, use the full module path, e.g., github.com/yourusername/yourprojectname)")
	fmt.Println("3. Tidy dependencies: go mod tidy")
	fmt.Println("   (This will download Fyne and other dependencies)")
	fmt.Printf("4. Run the application: go run ./cmd/%s/main.go\n", moduleName)
	fmt.Println("\nTo build an executable:")
	fmt.Printf("   go build -o %s ./cmd/%s/main.go\n", moduleName, moduleName)
	fmt.Printf("   Then run: ./%s (or %s.exe on Windows)\n", moduleName, moduleName)
}

