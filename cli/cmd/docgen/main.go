// Command docgen renders the CLI reference of the docs from the help texts in
// cmd/lucity, replacing everything between the cli-reference markers of a page.
package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	startMarker = "<!-- cli-reference:start"
	endMarker   = "<!-- cli-reference:end -->"
)

var tableColumns = map[string]string{
	"Commands":    "Command",
	"Arguments":   "Argument",
	"Flags":       "Flag",
	"Environment": "Variable",
}

var (
	sectionTitle = regexp.MustCompile(`^[A-Z][A-Za-z ]*:$`)
	codeToken    = regexp.MustCompile(`<[a-zA-Z][\w-]*>|--[a-z][a-z-]*|\b[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)+\b`)
)

type help struct {
	command string
	summary string
	blocks  []block
}

type block struct {
	title string
	lines []string
}

type entry struct {
	term        string
	description string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: docgen <package dir> <markdown page>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "docgen:", err)
		os.Exit(1)
	}
}

func run(packageDir, pagePath string) error {
	texts, err := usageTexts(packageDir)
	if err != nil {
		return err
	}

	var root *help
	commands := map[string]help{}
	for _, text := range texts {
		parsed, err := parseHelp(text)
		if err != nil {
			return err
		}
		if parsed.command == "" {
			root = &parsed
			continue
		}
		commands[parsed.command] = parsed
	}
	if root == nil {
		return fmt.Errorf("no top-level help text in %s", packageDir)
	}

	ordered, err := commandOrder(*root, commands)
	if err != nil {
		return err
	}
	return splice(pagePath, render(*root, ordered))
}

func usageTexts(packageDir string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(packageDir, "*.go"))
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	var files []*ast.File
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	config := types.Config{Importer: importer.ForCompiler(fileSet, "source", nil)}
	pkg, err := config.Check("main", fileSet, files, nil)
	if err != nil {
		return nil, fmt.Errorf("type-check %s: %w", packageDir, err)
	}

	var texts []string
	for _, name := range pkg.Scope().Names() {
		if name != "usage" && !strings.HasSuffix(name, "Usage") {
			continue
		}
		object, ok := pkg.Scope().Lookup(name).(*types.Const)
		if !ok || object.Val().Kind() != constant.String {
			continue
		}
		texts = append(texts, constant.StringVal(object.Val()))
	}
	return texts, nil
}

func parseHelp(text string) (help, error) {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	name, summary, ok := strings.Cut(lines[0], " — ")
	if !ok || (name != "lucity" && !strings.HasPrefix(name, "lucity ")) {
		return help{}, fmt.Errorf("help text must start with 'lucity <command> — <summary>', got %q", lines[0])
	}
	parsed := help{command: strings.TrimSpace(strings.TrimPrefix(name, "lucity")), summary: summary}

	previousBlank := false
	for _, line := range lines[1:] {
		last := len(parsed.blocks) - 1
		switch {
		case line == "":
			if last >= 0 && parsed.blocks[last].title != "" {
				parsed.blocks[last].lines = append(parsed.blocks[last].lines, "")
			}
			previousBlank = true
			continue
		case sectionTitle.MatchString(line):
			parsed.blocks = append(parsed.blocks, block{title: strings.TrimSuffix(line, ":")})
		case strings.HasPrefix(line, "  "):
			if last < 0 || parsed.blocks[last].title == "" {
				return help{}, fmt.Errorf("lucity %s: indented line outside a section: %q", parsed.command, line)
			}
			parsed.blocks[last].lines = append(parsed.blocks[last].lines, line[2:])
		default:
			if last < 0 || parsed.blocks[last].title != "" || previousBlank {
				parsed.blocks = append(parsed.blocks, block{})
				last = len(parsed.blocks) - 1
			}
			parsed.blocks[last].lines = append(parsed.blocks[last].lines, line)
		}
		previousBlank = false
	}

	for index := range parsed.blocks {
		lines := parsed.blocks[index].lines
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		parsed.blocks[index].lines = lines
	}
	return parsed, nil
}

func commandOrder(root help, commands map[string]help) ([]help, error) {
	var listed []entry
	for _, block := range root.blocks {
		if block.title == "Commands" {
			listed = entries(block.lines)
		}
	}

	var ordered []help
	seen := map[string]bool{}
	for _, item := range listed {
		name := strings.Fields(item.term)[0]
		command, ok := commands[name]
		if !ok {
			return nil, fmt.Errorf("command %q is listed in the top-level help but has no help text", name)
		}
		ordered = append(ordered, command)
		seen[name] = true
	}
	for name := range commands {
		if !seen[name] {
			return nil, fmt.Errorf("command %q has a help text but is missing from the top-level command list", name)
		}
	}
	return ordered, nil
}

func render(root help, commands []help) string {
	var out strings.Builder
	writeBlocks(&out, root)
	for _, command := range commands {
		fmt.Fprintf(&out, "### `%s`\n\n", command.command)
		fmt.Fprintf(&out, "%s.\n\n", strings.ToUpper(command.summary[:1])+command.summary[1:])
		writeBlocks(&out, command)
	}
	return out.String()
}

func writeBlocks(out *strings.Builder, parsed help) {
	for _, block := range parsed.blocks {
		column, isTable := tableColumns[block.title]
		switch {
		case block.title == "":
			writeParagraphs(out, block.lines)
		case block.title == "Usage":
			writeCode(out, "text", "", block.lines)
		case block.title == "Examples":
			writeCode(out, "bash", "Examples", block.lines)
		case isTable:
			writeTable(out, column, entries(block.lines), parsed.command == "" && block.title == "Commands")
		default:
			fmt.Fprintf(out, "**%s**\n\n", block.title)
			writeParagraphs(out, block.lines)
		}
	}
}

func writeParagraphs(out *strings.Builder, lines []string) {
	var paragraph []string
	flush := func() {
		if len(paragraph) > 0 {
			fmt.Fprintf(out, "%s\n\n", inline(strings.Join(paragraph, " ")))
			paragraph = nil
		}
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		paragraph = append(paragraph, strings.TrimSpace(line))
	}
	flush()
}

func writeCode(out *strings.Builder, language, label string, lines []string) {
	fence := "```" + language
	if label != "" {
		fence += " [" + label + "]"
	}
	fmt.Fprintf(out, "%s\n%s\n```\n\n", fence, strings.Join(lines, "\n"))
}

func writeTable(out *strings.Builder, column string, items []entry, linkToCommands bool) {
	fmt.Fprintf(out, "| %s | Description |\n| --- | --- |\n", column)
	for _, item := range items {
		term := "`" + strings.ReplaceAll(item.term, "|", `\|`) + "`"
		if linkToCommands {
			term = fmt.Sprintf("[%s](#%s)", term, strings.Fields(item.term)[0])
		}
		fmt.Fprintf(out, "| %s | %s |\n", term, strings.ReplaceAll(inline(item.description), "|", `\|`))
	}
	out.WriteString("\n")
}

func entries(lines []string) []entry {
	var items []entry
	for _, line := range lines {
		switch {
		case strings.TrimSpace(line) == "":
		case strings.HasPrefix(line, " ") && len(items) > 0:
			items[len(items)-1].description += " " + strings.TrimSpace(line)
		default:
			term, description, _ := strings.Cut(line, "  ")
			items = append(items, entry{term: term, description: strings.TrimSpace(description)})
		}
	}
	return items
}

func inline(text string) string {
	var out strings.Builder
	rest := text
	for {
		start, end := quotedSpan(rest)
		if start < 0 {
			out.WriteString(formatPlain(rest))
			return out.String()
		}
		out.WriteString(formatPlain(rest[:start]))
		out.WriteString("`" + rest[start+1:end] + "`")
		rest = rest[end+1:]
	}
}

func quotedSpan(text string) (int, int) {
	for start := 0; start < len(text); start++ {
		if text[start] != '\'' || (start > 0 && !strings.ContainsRune(" (", rune(text[start-1]))) {
			continue
		}
		for end := start + 2; end < len(text); end++ {
			if text[end] == '\'' && (end+1 == len(text) || strings.ContainsRune(" .,;:)", rune(text[end+1]))) {
				return start, end
			}
		}
	}
	return -1, -1
}

func formatPlain(text string) string {
	var out strings.Builder
	last := 0
	for _, match := range codeToken.FindAllStringIndex(text, -1) {
		out.WriteString(escapeHTML(text[last:match[0]]))
		out.WriteString("`" + text[match[0]:match[1]] + "`")
		last = match[1]
	}
	out.WriteString(escapeHTML(text[last:]))
	return out.String()
}

func escapeHTML(text string) string {
	return strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(text)
}

func splice(pagePath, reference string) error {
	data, err := os.ReadFile(pagePath)
	if err != nil {
		return err
	}
	page := string(data)

	start := strings.Index(page, startMarker)
	if start < 0 {
		return fmt.Errorf("%s has no %q marker", pagePath, startMarker)
	}
	bodyStart := start + strings.Index(page[start:], "\n") + 1
	end := strings.Index(page, endMarker)
	if end < bodyStart {
		return fmt.Errorf("%s has no %q marker after the start marker", pagePath, endMarker)
	}

	updated := page[:bodyStart] + "\n" + reference + page[end:]
	return os.WriteFile(pagePath, []byte(updated), 0o644)
}
