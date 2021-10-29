package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

const (
	DoneState   = "done"
	ActualState = "actual"

	GetOperation        = "get"
	AddOperation        = "add"
	RemoveOperation     = "rm"
	DoneOperation       = "done"
	UnexpectedOperation = "unexpected"

	StrikethroughChar = "\u0336"
	NewChar           = "🆕"
	DoneChar          = "✅"
	RemoveChar        = "🗑️"
	PooChar           = "💩"
	Pencil            = "📝"

	Help = "Unexpected format.\n" +
		"You can use:\n" +
		"1. \"todo\"\n" +
		"2. \"todo add your text\"\n" +
		"3. \"todo rm <ID>\"\n" +
		"4. \"todo done <ID>\"\n"
)

var (
	HomePath, _ = os.UserHomeDir()
	DBPath      = fmt.Sprintf("%s/.todo.json", HomePath)
)

type State string

func (s State) IsDone() bool {
	return s == DoneState
}

type Operation string

func (o Operation) isGet() bool {
	return o == GetOperation
}

func (o Operation) isAdd() bool {
	return o == AddOperation
}

func (o Operation) isRemove() bool {
	return o == RemoveOperation
}

func (o Operation) isDone() bool {
	return o == DoneOperation
}

func convertIDtoIdx(ID string) int {
	id, _ := strconv.Atoi(ID)
	return id - 1
}

func isDBExist() bool {
	_, err := os.Stat(DBPath)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

func updateDB(todos *TODOs) {
	_ = os.WriteFile(DBPath, todos.Marshal(), fs.ModePerm)
}

type TODO struct {
	Text  string `json:"text"`
	State State  `json:"state"`
}

func (td *TODO) StrikethroughText() string {
	return strings.Join(strings.Split(td.Text, ""), StrikethroughChar) + StrikethroughChar
}

type TODOs struct {
	All []TODO `json:"todos"`
}

func (tds *TODOs) Add(text string) string {
	tds.All = append(tds.All, TODO{
		Text:  text,
		State: ActualState,
	})
	updateDB(tds)
	return fmt.Sprintf("%s  The record with text \"%s\" has added!\n", NewChar, text)
}

func (tds *TODOs) Remove(ID string) string {
	response := fmt.Sprintf("%s  The record with ID \"%s\" has removed!\n", RemoveChar, ID)
	idx := convertIDtoIdx(ID)
	if idx < 0 || idx >= len(tds.All) {
		return response
	}
	tds.All = append(tds.All[:idx], tds.All[idx+1:]...)
	updateDB(tds)
	return response
}

func (tds *TODOs) Done(ID string) string {
	response := fmt.Sprintf("%s  The record with ID \"%s\" has done!\n", DoneChar, ID)
	idx := convertIDtoIdx(ID)
	if idx < 0 || idx >= len(tds.All) {
		return response
	}
	tds.All[idx].State = DoneState
	updateDB(tds)
	return response
}

func (tds *TODOs) String() string {
	result := fmt.Sprintf("%s  My TODO list:\n", Pencil)
	for idx, td := range tds.All {
		idx += 1
		if td.State.IsDone() {
			result += fmt.Sprintf("%d. %s\n", idx, td.StrikethroughText())
			continue
		}
		result += fmt.Sprintf("%d. %s\n", idx, td.Text)
	}
	return result
}

func (tds *TODOs) Unmarshal() {
	b, _ := os.ReadFile(DBPath)
	_ = json.Unmarshal(b, tds)
}

func (tds *TODOs) Marshal() []byte {
	b, _ := json.Marshal(tds)
	return b
}

func parseArgs(args []string) (Operation, string) {
	l := len(args)
	switch {
	case l == 0:
		return GetOperation, ""
	case l >= 2:
		op := Operation(args[0])
		if op.isAdd() {
			return AddOperation, strings.Join(args[1:], " ")
		}
		if op.isRemove() {
			return RemoveOperation, args[1]
		}
		if op.isDone() {
			return DoneOperation, args[1]
		}
	}
	return UnexpectedOperation, ""
}

func init() {
	if !isDBExist() {
		updateDB(new(TODOs))
	}
}

func main() {
	currentTODOs := new(TODOs)
	currentTODOs.Unmarshal()
	operation, value := parseArgs(os.Args[1:])
	switch operation {
	case GetOperation:
		fmt.Print(currentTODOs)
	case AddOperation:
		fmt.Print(currentTODOs.Add(value))
	case RemoveOperation:
		fmt.Print(currentTODOs.Remove(value))
	case DoneOperation:
		fmt.Print(currentTODOs.Done(value))
	default:
		fmt.Printf("%s  %s", PooChar, Help)
	}
}
