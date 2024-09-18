package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const TYPE_NIL byte = 0
const TYPE_NUMBER byte = 1
const TYPE_STRING byte = 2
const TYPE_FUNCTION byte = 3
const TYPE_LIST byte = 4
const TYPE_OTHER byte = 5
const TYPE_ERROR byte = 6

type VirtualMachine struct {
	stack     []any
	typestack []byte
	unused    []int
	temp      []int
	hashes    map[string]int
}

type Function func(vm *VirtualMachine, args []int) int

func init_vm() VirtualMachine {
	var vm VirtualMachine
	vm.hashes = make(map[string]int)
	return vm
}

func str_is_number(str string) bool {
	return regexp.MustCompile(`\d`).MatchString(str)
}

func specialSpaceSplit(str string) []string {
	var result []string
	var currentString strings.Builder
	recursion := 0
	insideQuotes := false

	for i, char := range str {
		switch char {
		case '(':
			if !insideQuotes {
				recursion++
			}
			currentString.WriteRune(char)
		case ')':
			if !insideQuotes {
				recursion--
			}
			currentString.WriteRune(char)
		case '"':
			insideQuotes = !insideQuotes
			currentString.WriteRune(char)
		default:
			if unicode.IsSpace(char) && recursion == 0 && !insideQuotes {
				if currentString.Len() > 0 {
					result = append(result, currentString.String())
					currentString.Reset()
				}
			} else {
				currentString.WriteRune(char)
			}
		}

		if i == len(str)-1 && currentString.Len() > 0 {
			result = append(result, currentString.String())
		}
	}

	return result
}

func specialSplit(str string, delim rune) []string {
	var result []string
	var currentString strings.Builder
	recursion := 0
	insideQuotes := false

	for i, char := range str {
		switch char {
		case '(':
			if !insideQuotes {
				recursion++
			}
			currentString.WriteRune(char)
		case ')':
			if !insideQuotes {
				recursion--
			}
			currentString.WriteRune(char)
		case '"':
			insideQuotes = !insideQuotes
			currentString.WriteRune(char)
		default:
			if char == delim && recursion == 0 && !insideQuotes {
				result = append(result, currentString.String())
				currentString.Reset()
			} else {
				currentString.WriteRune(char)
			}
		}

		if i == len(str)-1 {
			result = append(result, currentString.String())

		}
	}

	return result
}

func typeof(obj any) string {
	return fmt.Sprintf("%T", obj)
}

// constructors

func new_var(vm *VirtualMachine, value any, _type byte) int {
	var index int = -1

	if len(vm.unused) > 0 {
		index = vm.unused[0]
		vm.unused = vm.unused[1:]
		vm.typestack[index] = _type
		vm.stack[index] = value
	} else {
		vm.stack = append(vm.stack, value)
		vm.typestack = append(vm.typestack, _type)
		index = len(vm.stack) - 1
	}

	return index
}

func parse(vm *VirtualMachine, str string) []int {
	var splited []string = specialSpaceSplit(str)
	var result []int
	for _str := range splited {
		if strings.HasPrefix(splited[_str], "\"") {
			var _newstr string = strings.ReplaceAll(splited[_str], "\"", "")
			result = append(result, new_var(vm, _newstr, TYPE_STRING))
		} else if str_is_number(splited[_str]) {
			var v, error = strconv.ParseFloat(splited[_str], 64)
			if error != nil {
				panic(error)
			}
			result = append(result, new_var(vm, v, TYPE_NUMBER))
		} else if strings.Index(splited[_str], "(") != -1 {
			//expression & lists

		} else {
			result = append(result, vm.hashes[splited[_str]])
		}
	}
	return result
}

func interpret(vm *VirtualMachine, str string) int {
	var result int = -1
	var args []int = parse(vm, str)
	if vm.typestack[args[0]] == TYPE_FUNCTION { // run as func
		if _fn, ok := vm.stack[args[0]].(Function); ok {
			args = args[1:]
			result = (_fn)(vm, args)
		} else {
			result = -1
		}
	} else { // create a list
		var list []int
		for i := range args {
			list = append(list, args[i])
		}
		result = new_var(vm, list, TYPE_LIST)
	}
	return result
}

// std

func std_set(vm *VirtualMachine, args []int) int {
	var value int = args[1]

	if _varname, ok := vm.stack[args[0]].(string); ok {
		vm.hashes[_varname] = value
	}
	return -1
}

func std_print(vm *VirtualMachine, args []int) int {
	if _value, ok := vm.stack[args[0]].(Function); ok {
		println("{function}", _value)
	} else if _value, ok := vm.stack[args[0]].(string); ok {
		println("{string}", _value)
	} else if _value, ok := vm.stack[args[0]].(float64); ok {
		fmt.Printf("{number} %.5f\n", _value)
	} else if _value, ok := vm.stack[args[0]].([]int); ok {
		println("{list}", _value)
	} else {
		println("{unknown}")
	}
	return -1
}

func registerFunction(vm *VirtualMachine, functionName string, function Function) {
	var index int = new_var(vm, function, TYPE_FUNCTION)
	vm.hashes[functionName] = index
}

func main() {
	var vm VirtualMachine = init_vm()

	registerFunction(&vm, "set", std_set)
	registerFunction(&vm, "print", std_print)

	var index int = interpret(&vm, "set \"teste\" \"88\"")
	index = interpret(&vm, "print teste")
	println(index)
}
