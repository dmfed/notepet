package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/dmfed/notepet"
	"github.com/dmfed/termtools"
)

var prnt = termtools.PrintSuite{}

var (
	noteTitleMarker  string = ">>>"
	noteStickyMarker        = "::s::"
	noteTagsStart           = ">"
	noteTagsEnd             = "<"
)

// regexes to lookup title, tags and sticky attribute
var (
	noteTitleRe  = regexp.MustCompile(prnt.Sprintf(`\n*%v *(.+)\n`, noteTitleMarker))
	noteStickyRe = regexp.MustCompile(prnt.Sprintf(`\n+%v\n*`, noteStickyMarker))
	noteTagsRe   = regexp.MustCompile(prnt.Sprintf(`\n+%v(.+)%v\n*`, noteTagsStart, noteTagsEnd))
)

// default colors to set up termtools printers
var (
	printerDefaultNoteHeaderColor     int = 3
	printerDefaultNoteStickyAttrColor     = 9
	// printerDefaultNoteTagsColor           = 33
)

// new commands can be implemented by writing a function and adding it to this map
var (
	knownCommands = map[string]func(notepet.Storage, *notepetConfig) error{
		"show":   processShowCommand,
		"put":    processPutCommand,
		"new":    processNewCommand,
		"sticky": processStickyCommand,
		"del":    processDelCommand,
		"edit":   processEditCommand,
		"search": processSearchCommand,
		"export": processExportCommand,
		"shell":  processShellCommand,
	}
)

// runCLI processes input from command line.
// Note that this function relies on flag.Parse()
// being called before it (we're calling flag.Arg() here).
func RunCLI(st notepet.Storage, conf *notepetConfig, cconf notepet.ClientConfig) error {
	if cconf.Color.Enabled {
		setupPrinters()
	}
	command := strings.ToLower(flag.Arg(0))
	processor, ok := knownCommands[command]
	if !ok {
		// displayHelpLong()
		return prnt.Use("error").Errorf("error: unrecognized command supplied: %v", command)
	}
	return processor(st, conf)
}

func processShowCommand(st notepet.Storage, conf *notepetConfig) error {
	notes, err := st.Get()
	if len(notes) == 0 {
		return prnt.Errorf("no notes found: %v", err)
	}
	start, end, err := parseSliceArg(flag.Arg(1), len(notes))
	if err != nil {
		return prnt.Use("error").Errorf("%v", err)
	}
	notes = notes[start:end]
	for _, note := range notes {
		prnt.Use("header").Print(start+1, " ")
		// note.ID = notepet.NoteID(prnt.Sprintf("%v", start+1))
		printNote(note, conf)
		start++
	}
	return nil
}

func processPutCommand(st notepet.Storage, conf *notepetConfig) error {
	note := notepet.Note{}
	if flag.Arg(3) != "" {
		note.Tags = flag.Arg(3)
	}
	if flag.Arg(2) != "" {
		note.Body = flag.Arg(2)
		note.Title = flag.Arg(1)
	} else if flag.Arg(1) != "" {
		note.Body = flag.Arg(1)
	}
	id, err := st.Put(note)
	if err == nil {
		prnt.Printf("Added note. New note ID is: %v\n", id)
	}
	return err
}

func processNewCommand(st notepet.Storage, conf *notepetConfig) error {
	note, err := editNewNote(conf)
	if err != nil {
		prnt.Println("Failed to create new note.")
		return err
	}
	prnt.Println("Created new note.")
	id, err := st.Put(note)
	if err == nil {
		prnt.Printf("Saved note to Storage. New note ID is: %v\n", id)
	}
	return err
}

func processStickyCommand(st notepet.Storage, conf *notepetConfig) error {
	index, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return prnt.Errorf("invalid index")
	}
	notes, _ := st.Get()
	if index-1 < 0 || index-1 >= len(notes) {
		return prnt.Errorf("invalid index")
	}
	note := notes[index-1]
	note.Sticky = !note.Sticky
	id, err := st.Upd(note.ID, note)
	if err == nil {
		prnt.Printf("Set STICKY mode for ID %v to %v\n", id, note.Sticky)
	}
	return err
}

func processDelCommand(st notepet.Storage, conf *notepetConfig) error {
	index, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return prnt.Errorf("invalid index")
	}
	notes, _ := st.Get()
	if index-1 < 0 || index-1 >= len(notes) {
		return prnt.Errorf("invalid index")
	}
	note := notes[index-1]
	printNote(note, conf)
	if !promptUserYorN("Delete this note?") {
		return nil
	}
	err = st.Del(note.ID)
	if err == nil {
		prnt.Printf("Successfully deleted note with id %v\n", note.ID)
	}
	return err
}

func processEditCommand(st notepet.Storage, conf *notepetConfig) error {
	index, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		return prnt.Errorf("invalid index")
	}
	notes, _ := st.Get()
	if index-1 < 0 || index-1 >= len(notes) {
		return prnt.Errorf("invalid index")
	}
	note := notes[index-1]
	oldID := note.ID
	printNote(note, conf)
	if !promptUserYorN("Edit this note?") {
		return nil
	}
	note, err = editNote(note, conf)
	if err != nil {
		prnt.Println("Could not edit note.")
		return err
	}
	prnt.Println("Sucessfully edited note.")
	newID, err := st.Upd(oldID, note)
	if err == nil {
		prnt.Printf("Sucessfully replaced note with id %v.\n", newID)
	}
	return err
}

func processSearchCommand(st notepet.Storage, conf *notepetConfig) error {
	stringToFind := flag.Arg(1)
	notes, err := st.Search(stringToFind)
	if err != nil {
		return err
	}
	for _, note := range notes {
		printNote(note, conf)
	}
	return nil
}

func processExportCommand(st notepet.Storage, conf *notepetConfig) error {
	data, err := notepet.ExportJSON(st)
	if err == nil {
		prnt.Println(string(data))
	}
	return err
}

func processShellCommand(st notepet.Storage, conf *notepetConfig) error {
	termtools.ClearScreen()

	return nil
}

func editNewNote(conf *notepetConfig) (note notepet.Note, err error) {
	note.Title = " "
	note.Tags = " "
	note.Sticky = true
	return editNote(note, conf)
}

func editNote(n notepet.Note, conf *notepetConfig) (note notepet.Note, err error) {
	tmpFile, err := createTempFile()
	if err != nil {
		return
	}
	defer os.Remove(tmpFile.Name())
	// defer tmpFile.Close()
	// TODO: Handle errors here ?
	notebytes := []byte(convertNoteToEditableString(n))
	tmpFile.Write(notebytes)
	tmpFile.Close() // Closing and reopening (keep file open and use Seek(0, 0)?)
	err = runEditor(tmpFile.Name(), conf.editor)
	if err != nil {
		return
	}
	tmpFile, err = os.Open(tmpFile.Name())
	if err != nil {
		return
	}
	defer tmpFile.Close()
	data, err := ioutil.ReadAll(tmpFile)
	if err != nil {
		return
	}
	return convertStringToNote(string(data)), nil
}

func convertNoteToEditableString(n notepet.Note) (output string) {
	if n.Title != "" {
		output += noteTitleMarker + n.Title + "\n"
	}
	output += n.Body + "\n"
	if n.Sticky {
		output += noteStickyMarker + "\n"
	}
	if n.Tags != "" {
		output += noteTagsStart + n.Tags + noteTagsEnd + "\n"
	}
	return
}

func convertStringToNote(input string) (note notepet.Note) {
	if noteTitleRe.MatchString(input) {
		note.Title = noteTitleRe.FindStringSubmatch(input)[1]
		loc := noteTitleRe.FindStringIndex(input)
		input = input[:loc[0]] + input[loc[1]:]
	}
	if noteTagsRe.MatchString(input) {
		note.Tags = noteTagsRe.FindStringSubmatch(input)[1]
		loc := noteTagsRe.FindStringIndex(input)
		input = input[:loc[0]] + input[loc[1]:]
	}
	if noteStickyRe.MatchString(input) {
		note.Sticky = true
		loc := noteStickyRe.FindStringIndex(input)
		input = input[:loc[0]] + input[loc[1]:]
	} else {
		note.Sticky = false
	}
	note.Body = strings.TrimRight(input, " \n")
	return
}

func runEditor(filename, editorcommand string) error {
	editor := exec.Command(editorcommand, filename)
	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	return editor.Run()
}

func createTempFile() (file *os.File, err error) {
	homedir, _ := os.UserHomeDir()
	file, err = os.CreateTemp(homedir, ".notepet*.swp")
	return
}

func printNote(note notepet.Note, conf *notepetConfig) {
	switch conf.verbose {
	case true:
		printNoteVerbose(note)
	default:
		printNoteRegular(note)
	}
}

func printNoteVerbose(note notepet.Note) {
	var out string
	out += prnt.Sprintf("ID:\t\t%v\n", note.ID.String())
	out += prnt.Sprintf("Taken:\t\t%v\n", note.TimeStamp.Format("02/01/2006 15:04:05"))
	out += prnt.Sprintf("Edited:\t\t%v ", note.LastEdited.Format("02/01/2006 15:04:05"))
	if note.Sticky {
		out += prnt.Use("sticky").Sprint("STICKY") + "\n"
	} else {
		out += "\n"
	}
	if note.Title != "" {
		out += prnt.Sprint("Title:\t\t") + prnt.Use("header").Sprint(note.Title) + "\n"
	}
	out += prnt.Use("body").Sprint(note.Body) + "\n"
	if note.Tags != "" {
		out += prnt.Sprint("Tags:\t\t") + prnt.Use("tags").Sprint(note.Tags) + "\n"
	}
	prnt.Println(out)
}

func printNoteRegular(note notepet.Note) {
	var out string
	if note.Sticky {
		out += prnt.Use("sticky").Sprint("STICKY") + " "
	}
	if note.Title != "" {
		out += prnt.Use("header").Sprintln(note.Title)
	}
	out += prnt.Use("header").Sprintf("%v\n", note.LastEdited.Format("02/01/2006 15:04:05"))
	out += prnt.Use("body").Sprintln(note.Body)
	if note.Tags != "" {
		out += prnt.Use("tags").Sprintf("%v %v %v\n", noteTagsStart, note.Tags, noteTagsEnd)
	}
	prnt.Println(out)
}

func parseSliceArg(input string, maxindex int) (start, end int, err error) {
	indices := strings.Split(input, ":")
	// allow Python-style slicing in 1-based [x,y) index
	if len(indices) > 1 {
		if num, err := strconv.Atoi(indices[0]); err == nil {
			start = num - 1 // displayed indexes start with 1, so 1 stands for 0 in actual slice
		}
		if num, err := strconv.Atoi(indices[1]); err == nil {
			end = num - 1 // if user wants a slice, last index is not included
		}
	} else {
		if num, err := strconv.Atoi(indices[0]); err == nil {
			start, end = num-1, num // will yield exact index of one note when we range over resulting indexes
		}
	}
	if start < 0 { // allow Python style slicing e.g. user supplied "-5:2"
		start = maxindex + start + 1
	}
	if end < 0 { // e.g. user supplied "1:-3"
		end = maxindex + end + 1
	}
	if start >= end && end == 0 { // e.g. user supplied "3:" or "-3:"
		end = maxindex
	}
	if start < 0 || start >= maxindex || start >= end || end < start+1 || end > maxindex {
		err = prnt.Errorf("error: index out of bounds %v:%v with total %v notes", start, end, maxindex)
		start, end = 0, 0 // fallback to safe values to make sure program doesn't crash if they are used
		return
	}
	return
}

func promptUserYorN(question string) (result bool) {
	var answer string
	for {
		prnt.Print(question, " y/n: ")
		fmt.Scan(&answer)
		answer = strings.ToLower(answer)
		switch answer {
		case "y", "yes":
			result = true
			return
		case "n", "no":
			return
		default:
			prnt.Println("\nPlease type \"yes\" or \"no\"...")
		}
	}
}

func setupPrinters() {
	configs := []termtools.PrinterConfig{
		{Name: "header", Color: printerDefaultNoteHeaderColor},
		{Name: "sticky", Color: printerDefaultNoteStickyAttrColor, Reversed: true},
		{Name: "tags", Bold: true},
		{Name: "body", Color: "black"},
		{Name: "error", Color: "red"}}
	prnt.Configure(configs...)
}
