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
