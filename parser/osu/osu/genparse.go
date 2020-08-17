// +build ignore

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// too many inconsistent pattern in .osu file format, very hard to write fully-generating code
type fieldInfo struct {
	name      string
	fieldType string
	delimiter []string
}

// ScanStructs supposes gofmt at given file was already proceeded
func ScanStructs(path string) ([]string, map[string]string, map[string][]fieldInfo) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	structs := make([]string, 0)
	delimiters := make(map[string]string)
	m := make(map[string][]fieldInfo)
	scanner := bufio.NewScanner(f)
	var structName string
	var infos []fieldInfo
	for scanner.Scan() {
		vs := strings.Fields(scanner.Text())
		switch {
		case len(vs) == 0 || vs[0] == "//" || vs[len(vs)-1] == "manual": // maybe panic won't happen
			continue
		case vs[0] == "type" && len(vs) > 2 && vs[2] == "struct":
			structName = vs[1]
			structs = append(structs, structName)
			if strings.HasPrefix(vs[len(vs)-1], `delimiter`) {
				delimiter := strings.TrimLeft(vs[len(vs)-1], `delimiter`)
				delimiters[structName] = delimiter
			}
			infos = make([]fieldInfo, 0)
		case structName != "" && len(vs) >= 2:
			info := fieldInfo{name: vs[0], fieldType: vs[1]}
			info.delimiter = make([]string, 0)
			for i := 0; i < strings.Count(vs[1], "["); i++ {
				delimiter := strings.TrimLeft(vs[3+2*i], `delimiter`)
				if delimiter == "(space)" {
					delimiter = " "
				}
				info.delimiter = append(info.delimiter, delimiter)
			}
			infos = append(infos, info)
		case vs[0] == "}":
			m[structName] = infos
			structName = ""
		}
	}
	return structs, delimiters, m
}

// todo: make it easier to read
func PrintSetValue(fields []fieldInfo, structName, delimiter, genMode string) {
	var localName, ptrmark, returnName, valName string
	delimiter = strings.Replace(delimiter, "(space)", " ", -1)
	switch genMode {
	case "section":
		localName = "o." + structName
		ptrmark = "&"
		returnName = "o"
		valName = "kv[1]"

		fmt.Printf("case \"%s\":\n", structName)
		fmt.Printf("kv := strings.Split(line, `%s`)\n", delimiter)
		fmt.Printf("switch kv[0] {\n")
	case "substruct":
		localName = genLocalName(structName)
		ptrmark = ""
		returnName = localName

		fmt.Printf("\nfunc new%s(line string) (%s, error) {\n", structName, structName)
		fmt.Printf("var %s %s\n", localName, structName)
	}
	for i, f := range fields {
		switch genMode {
		case "section":
			fmt.Printf("case \"%s\":", f.name)
		case "substruct":
			valName = fmt.Sprintf("v[%d]", i)
			fmt.Printf("{") // block
		}
		switch f.fieldType {
		case "string":
			fmt.Printf(`
	%s.%s = %s
`, localName, f.name, valName)
		case "int":
			fmt.Printf(`
	i, err := strconv.Atoi(%s)
	if err != nil {
			return %s%s, err
		}
	%s.%s = i
`, valName, ptrmark, returnName, localName, f.name)
		case "float64":
			fmt.Printf(`
	f, err := strconv.ParseFloat(%s, 64)
	if err != nil {
			return %s%s, err
		}
	%s.%s = f
`, valName, ptrmark, returnName, localName, f.name)
		case "bool":
			fmt.Printf(`
	b, err := strconv.ParseBool(%s)
	if err != nil {
		return %s%s, err
		}
	%s.%s = b
`, valName, ptrmark, returnName, localName, f.name)
		case "[]string":
			fmt.Printf(`
	slice := make([]string, 0)
	for _, s := range strings.Split(%s, "%s") {
		slice = append(slice, s)
	}
	%s.%s = slice
`, valName, f.delimiter[0], localName, f.name)
		case "[]int":
			fmt.Printf(`
	slice := make([]int, 0)
	for _, s := range strings.Split(%s, "%s") {
		i, err := strconv.Atoi(s)
		if err != nil {
			return %s%s, err
		}
		slice = append(slice, i)
	}
	%s.%s = slice
`, valName, f.delimiter[0], ptrmark, returnName, localName, f.name)
		}
		switch genMode {
		case "substruct":
			fmt.Printf("}\n") // block
		}
	}
	switch genMode {
	case "section":
		fmt.Printf("}\n")
	case "substruct":
		fmt.Printf(`
return %s, nil
}`, localName)
	}
}

// generate local name
// example: TimingPoint -> tp
func genLocalName(structName string) string {
	var name string
	for i, s := range strings.ToLower(structName) {
		if structName[i] != byte(s) {
			name += string(s)
		}
	}
	return name
}

func main() {
	structs, delimiters, m := ScanStructs("format.go")
	for _, s := range structs {
		switch s {
		case "General", "Editor", "Metadata", "Difficulty":
			PrintSetValue(m[s], s, delimiters[s], "section")
		case "TimingPoint", "HitObject", "SliderParams", "HitSample":
			PrintSetValue(m[s], s, delimiters[s], "substruct")
		}
		// fmt.Printf("%s: %q\n", s, m[s])
	}
}