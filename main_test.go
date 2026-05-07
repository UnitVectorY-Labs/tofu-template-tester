package main

import (
   "os"
   "path/filepath"
   "testing"
)

// Helper to compare two string slices for equality
func equalStringSlices(a, b []string) bool {
   if len(a) != len(b) {
       return false
   }
   for i := range a {
       if a[i] != b[i] {
           return false
       }
   }
   return true
}

func TestListParams(t *testing.T) {
   tmpl := "Hello ${USER}, your ID is ${ ID } and again ${USER}"
   want := []string{"ID", "USER"}
   got := listParams(tmpl)
   if !equalStringSlices(got, want) {
       t.Errorf("listParams(%q) = %v; want %v", tmpl, got, want)
   }
}

func TestLoadProperties_Valid(t *testing.T) {
   dir := t.TempDir()
   path := filepath.Join(dir, "props.txt")
   content := `# comment
FOO=bar
BAZ = qux

` // valid properties
   if err := os.WriteFile(path, []byte(content), 0644); err != nil {
       t.Fatalf("failed to write temp file: %v", err)
   }
   props, err := loadProperties(path)
   if err != nil {
       t.Fatalf("loadProperties returned error: %v", err)
   }
   want := map[string]string{"FOO": "bar", "BAZ": "qux"}
   if len(props) != len(want) {
       t.Fatalf("expected %d properties; got %d", len(want), len(props))
   }
   for k, v := range want {
       if props[k] != v {
           t.Errorf("props[%q] = %q; want %q", k, props[k], v)
       }
   }
}

func TestLoadProperties_InvalidLine(t *testing.T) {
   dir := t.TempDir()
   path := filepath.Join(dir, "invalid.txt")
   content := "INVALID_LINE\n"
   if err := os.WriteFile(path, []byte(content), 0644); err != nil {
       t.Fatalf("failed to write temp file: %v", err)
   }
   _, err := loadProperties(path)
   if err == nil {
       t.Fatal("expected error for invalid properties line; got nil")
   }
}

func TestListParams_NoVars(t *testing.T) {
   tmpl := "Hello world, no variables here"
   got := listParams(tmpl)
   if len(got) != 0 {
       t.Errorf("listParams(%q) = %v; want empty", tmpl, got)
   }
}

func TestListParams_EmptyTemplate(t *testing.T) {
   got := listParams("")
   if len(got) != 0 {
       t.Errorf("listParams(\"\") = %v; want empty", got)
   }
}

func TestListParams_EscapedDollar(t *testing.T) {
   // $${VAR} is an escape sequence; VAR should NOT appear in the list
   tmpl := "echo $${HOME} and ${NAME}"
   want := []string{"NAME"}
   got := listParams(tmpl)
   if !equalStringSlices(got, want) {
       t.Errorf("listParams(%q) = %v; want %v", tmpl, got, want)
   }
}

func TestListParams_OnlyEscaped(t *testing.T) {
   tmpl := "$${FOO} and $${BAR}"
   got := listParams(tmpl)
   if len(got) != 0 {
       t.Errorf("listParams(%q) = %v; want empty", tmpl, got)
   }
}

func TestListParams_PercentEscape(t *testing.T) {
   // %%{ is an escape for %{; should not affect variable listing
   tmpl := "%%{if true}${VAR}%%{endif}"
   want := []string{"VAR"}
   got := listParams(tmpl)
   if !equalStringSlices(got, want) {
       t.Errorf("listParams(%q) = %v; want %v", tmpl, got, want)
   }
}

func TestListParams_Underscore(t *testing.T) {
   tmpl := "${my_var} and ${MY_VAR_2}"
   want := []string{"MY_VAR_2", "my_var"}
   got := listParams(tmpl)
   if !equalStringSlices(got, want) {
       t.Errorf("listParams(%q) = %v; want %v", tmpl, got, want)
   }
}

func TestLoadProperties_ValueWithEquals(t *testing.T) {
   dir := t.TempDir()
   path := filepath.Join(dir, "props.txt")
   content := "KEY=val=ue\n"
   if err := os.WriteFile(path, []byte(content), 0644); err != nil {
       t.Fatalf("failed to write temp file: %v", err)
   }
   props, err := loadProperties(path)
   if err != nil {
       t.Fatalf("loadProperties returned error: %v", err)
   }
   if props["KEY"] != "val=ue" {
       t.Errorf("props[KEY] = %q; want %q", props["KEY"], "val=ue")
   }
}

func TestLoadProperties_EmptyValue(t *testing.T) {
   dir := t.TempDir()
   path := filepath.Join(dir, "props.txt")
   content := "KEY=\n"
   if err := os.WriteFile(path, []byte(content), 0644); err != nil {
       t.Fatalf("failed to write temp file: %v", err)
   }
   props, err := loadProperties(path)
   if err != nil {
       t.Fatalf("loadProperties returned error: %v", err)
   }
   if props["KEY"] != "" {
       t.Errorf("props[KEY] = %q; want empty string", props["KEY"])
   }
}

func TestLoadProperties_OnlyComments(t *testing.T) {
   dir := t.TempDir()
   path := filepath.Join(dir, "props.txt")
   content := "# comment 1\n# comment 2\n\n"
   if err := os.WriteFile(path, []byte(content), 0644); err != nil {
       t.Fatalf("failed to write temp file: %v", err)
   }
   props, err := loadProperties(path)
   if err != nil {
       t.Fatalf("loadProperties returned error: %v", err)
   }
   if len(props) != 0 {
       t.Errorf("expected 0 properties; got %d", len(props))
   }
}

func TestLoadProperties_FileNotFound(t *testing.T) {
   _, err := loadProperties("/nonexistent/path/file.txt")
   if err == nil {
       t.Fatal("expected error for missing file; got nil")
   }
}

func TestProcessTemplate_Success(t *testing.T) {
   tmpl := "A:${A}, B: ${B}"
   props := map[string]string{"A": "1", "B": "2"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "A:1, B: 2"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_MissingProperty(t *testing.T) {
   tmpl := "Hello ${NAME}"
   props := map[string]string{}
   _, err := processTemplate(tmpl, props)
   if err == nil {
       t.Fatal("expected error for missing property; got nil")
   }
}

func TestProcessTemplate_EmptyTemplate(t *testing.T) {
   out, err := processTemplate("", map[string]string{})
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   if out != "" {
       t.Errorf("processTemplate output = %q; want empty string", out)
   }
}

func TestProcessTemplate_NoVars(t *testing.T) {
   tmpl := "Hello world, no variables"
   out, err := processTemplate(tmpl, map[string]string{})
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   if out != tmpl {
       t.Errorf("processTemplate output = %q; want %q", out, tmpl)
   }
}

func TestProcessTemplate_DollarEscape(t *testing.T) {
   // $${HOME} should produce literal ${HOME}, not substitute
   tmpl := "echo $${HOME}"
   props := map[string]string{"HOME": "/root"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "echo ${HOME}"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_DollarEscapeMixed(t *testing.T) {
   // Mix of escaped and regular variables
   tmpl := "literal: $${ESCAPED}, replaced: ${NAME}"
   props := map[string]string{"NAME": "alice"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "literal: ${ESCAPED}, replaced: alice"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_PercentEscape(t *testing.T) {
   // %%{ should produce literal %{
   tmpl := "%%{if true}content%%{endif}"
   out, err := processTemplate(tmpl, map[string]string{})
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "%{if true}content%{endif}"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_WhitespaceInVar(t *testing.T) {
   tmpl := "Hello ${ NAME }"
   props := map[string]string{"NAME": "world"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "Hello world"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_SameVarMultiple(t *testing.T) {
   tmpl := "${X} and ${X} again"
   props := map[string]string{"X": "val"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "val and val again"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_SpecialCharsInValue(t *testing.T) {
   tmpl := "conn=${CONN}"
   props := map[string]string{"CONN": "host=localhost port=5432 user=$admin"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "conn=host=localhost port=5432 user=$admin"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_ValueContainsDollarBrace(t *testing.T) {
   // Value containing ${...} should be inserted literally (not re-processed)
   tmpl := "result=${VAL}"
   props := map[string]string{"VAL": "${NOT_A_VAR}"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "result=${NOT_A_VAR}"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_LoneDollar(t *testing.T) {
   // A lone $ not followed by { should pass through
   tmpl := "price is $10"
   out, err := processTemplate(tmpl, map[string]string{})
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   if out != tmpl {
       t.Errorf("processTemplate output = %q; want %q", out, tmpl)
   }
}

func TestProcessTemplate_TripleDollarEscape(t *testing.T) {
   // $$${NAME}: first $ is literal, then $${ is the escape producing ${, then NAME} is literal
   tmpl := "$$${NAME}"
   props := map[string]string{"NAME": "test"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "$${NAME}"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestProcessTemplate_MultilineTemplate(t *testing.T) {
   tmpl := "line1: ${A}\nline2: ${B}\nline3: plain"
   props := map[string]string{"A": "val1", "B": "val2"}
   out, err := processTemplate(tmpl, props)
   if err != nil {
       t.Fatalf("processTemplate returned error: %v", err)
   }
   want := "line1: val1\nline2: val2\nline3: plain"
   if out != want {
       t.Errorf("processTemplate output = %q; want %q", out, want)
   }
}

func TestReadInput_FromFile(t *testing.T) {
   dir := t.TempDir()
   path := filepath.Join(dir, "input.txt")
   content := "Hello ${NAME}"
   if err := os.WriteFile(path, []byte(content), 0644); err != nil {
       t.Fatalf("failed to write temp file: %v", err)
   }
   got, err := readInput(path)
   if err != nil {
       t.Fatalf("readInput returned error: %v", err)
   }
   if got != content {
       t.Errorf("readInput = %q; want %q", got, content)
   }
}

func TestReadInput_FileNotFound(t *testing.T) {
   _, err := readInput("/nonexistent/path/file.txt")
   if err == nil {
       t.Fatal("expected error for missing file; got nil")
   }
}

func TestWriteOutput_ToFile(t *testing.T) {
   dir := t.TempDir()
   path := filepath.Join(dir, "output.txt")
   content := "Hello world"
   if err := writeOutput(content, path); err != nil {
       t.Fatalf("writeOutput returned error: %v", err)
   }
   got, err := os.ReadFile(path)
   if err != nil {
       t.Fatalf("failed to read output file: %v", err)
   }
   if string(got) != content {
       t.Errorf("writeOutput file content = %q; want %q", string(got), content)
   }
}

func TestWriteOutput_InvalidPath(t *testing.T) {
   err := writeOutput("content", "/nonexistent/dir/file.txt")
   if err == nil {
       t.Fatal("expected error for invalid output path; got nil")
   }
}
