package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Arguments map[string]string

var flags = []string{"operation", "id", "item", "fileName"}

func Perform(args Arguments, writer io.Writer) error {
	if args[flags[0]] == "" {
		return EmptyFlagError(flags[0])
	}
	if args[flags[3]] == "" {
		return EmptyFlagError(flags[3])
	}
	switch args[flags[0]] {
	case "list":
		return List(args, writer)
	case "add":
		return Add(args, writer)
	case "remove":
		return RemoveById(args, writer)
	case "findById":
		return FindById(args, writer)
	default:
		return fmt.Errorf("Operation %s not allowed!", args[flags[0]])
	}
}

func EmptyFlagError(flag string) error {
	return fmt.Errorf("-%s flag has to be specified", flag)
}

func List(args Arguments, writer io.Writer) error {
	file, err := os.OpenFile(args[flags[3]], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	str, err := io.ReadAll(file)
	_, err = writer.Write(str)
	return err
}

func Add(args Arguments, writer io.Writer) error {
	if args["item"] == "" {
		return EmptyFlagError("item")
	}
	file, err := os.OpenFile(args[flags[3]], os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	str, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	obj := struct{ Id string }{}
	err = json.Unmarshal([]byte(args["item"]), &obj)
	if err != nil {
		panic(err)
	}
	if strings.Contains(string(str), fmt.Sprintf("\"id\":\"%s\"", obj.Id)) {
		_, err := writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", obj.Id)))
		if err != nil {
			return err
		}
		return nil
	}
	
	if len(str) == 0 {
		_, err := file.Write([]byte("[" + args[flags[2]] + "]"))
		if err != nil {
			return err
		}
	} else {
		_, err = file.WriteAt([]byte(fmt.Sprintf(",%s]", args[flags[2]])), int64(len(str)-1))
		if err != nil {
			return err
		}
	}
	return nil
}

func FindObjById(id string, txt string) (string, int) {
	start := strings.Index(txt, fmt.Sprintf("\"id\":\"%s\"", id))
	if start < 0 {
		return "", start
	}
	var result string
	for i := start; ; i++ {
		if string(txt[i]) == "}" {
			result = txt[start-1 : i+1]
			break
		}
	}
	return result, start - 1
}

func FindById(args Arguments, writer io.Writer) error {
	id := args[flags[1]]
	if id == "" {
		return EmptyFlagError(flags[1])
	}
	file, err := os.OpenFile(args[flags[3]], os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	txt, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	user, _ := FindObjById(id, string(txt))
	_, err = writer.Write([]byte(user))
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	return nil
}

func RemoveById(args Arguments, writer io.Writer) error {
	id := args[flags[1]]
	if id == "" {
		return EmptyFlagError(flags[1])
	}
	file, err := os.OpenFile(args[flags[3]], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	txt, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	user, i := FindObjById(id, string(txt))
	end := i + len(user)
	if user == "" {
		_, err = writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
		if err != nil {
			return err
		}
	} else {
		if string(txt[end]) == "," {
			txt = append(txt[:i], txt[end+1:]...)
		} else if string(txt[i-1]) == "," {
			txt = append(txt[:i-1], txt[end:]...)
		} else {
			txt = []byte{}
		}
		err = file.Truncate(0)
		_, err = file.WriteAt(txt, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	
	if err != nil {
		panic(err)
	}
}

func parseArgs() Arguments {
	res := Arguments{
		"id":        "",
		"item":      "",
		"operation": "",
		"fileName":  "",
	}
	args := os.Args[1:]
	for i, v := range args {
		if strings.HasPrefix(v, "-") {
			if flag := strings.TrimLeft(v, "-"); isFlag(flag) && i+1 != len(args) {
				res[flag] = args[i+1]
			}
		}
	}
	obj := struct{ Id string }{}
	err := json.Unmarshal([]byte(res["item"]), &obj)
	if err != nil {
		panic(err)
	}
	res["id"] = obj.Id
	return res
}

func isFlag(str string) bool {
	for _, flag := range flags {
		if str == flag {
			return true
		}
	}
	return false
}
