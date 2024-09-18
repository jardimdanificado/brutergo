package main

import (
	"fmt"
	"os"
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

func removeIntElement(obj []int, index int) []int {
	for i := index; i < len(obj)-1; i++ {
		obj[i] = obj[i+1]
	}
	obj = obj[:len(obj)-1]
	return obj
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

	vm.temp = append(vm.temp, index)

	return index
}

func parse(vm *VirtualMachine, str string) []int {
	var splited []string = specialSpaceSplit(str)
	var result []int
	for _str := range splited {
		if strings.HasPrefix(splited[_str], "\"") {
			var _newstr string = splited[_str][1 : len(splited[_str])-1]
			result = append(result, new_var(vm, _newstr, TYPE_STRING))
		} else if str_is_number(splited[_str]) {
			var v, error = strconv.ParseFloat(splited[_str], 64)
			if error != nil {
				panic(error)
			}
			result = append(result, new_var(vm, v, TYPE_NUMBER))
		} else if strings.Index(splited[_str], "(") != -1 {
			interpret(vm, splited[_str][1:len(splited[_str])-1])
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

func eval(vm *VirtualMachine, str string) int {
	var splited []string = specialSplit(str, ';')
	var result int = -1
	for i := range splited {
		result = interpret(vm, splited[i])
		if result != -1 {
			break
		}
	}
	return result
}

// std
// std
// std
// std
// std
// std
// std

func std_set(vm *VirtualMachine, args []int) int {
	var value int = args[1]

	if _varname, ok := vm.stack[args[0]].(string); ok {
		vm.hashes[_varname] = value
	}
	return -1
}

func std_rm(vm *VirtualMachine, args []int) int {
	vm.typestack[args[0]] = TYPE_NIL
	vm.stack[args[0]] = 0
	vm.unused = append(vm.unused, args[0])
	return -1
}

func std_clear(vm *VirtualMachine, args []int) int { //need to fix
	for i := range vm.temp {
		vm.typestack[vm.temp[i]] = TYPE_NIL
		vm.stack[vm.temp[i]] = -1
		vm.unused = append(vm.unused, vm.temp[i])
	}
	vm.temp = vm.temp[len(vm.temp):]
	return -1
}

func std_edit(vm *VirtualMachine, args []int) int {

	vm.stack[args[0]] = vm.stack[args[1]]
	vm.typestack[args[0]] = vm.typestack[args[1]]

	return -1
}

func std_hold(vm *VirtualMachine, args []int) int { // need to fix
	for i := 0; i < len(vm.temp); i++ {
		if vm.temp[i] == args[0] {
			vm.temp = removeIntElement(vm.temp, i)
		}
	}
	return -1
}

func std_unhold(vm *VirtualMachine, args []int) int {
	vm.temp = append(vm.temp, args[0])
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

// main
// main
// main
// main
// main
// main
// main

func registerFunction(vm *VirtualMachine, functionName string, function Function) {
	var index int = new_var(vm, function, TYPE_FUNCTION)
	vm.hashes[functionName] = index
}

func main() {
	var vm VirtualMachine = init_vm()

	var _txt, error = os.ReadFile(os.Args[1])
	if error != nil {
		panic("file not found.")
	}

	registerFunction(&vm, "set", std_set)
	registerFunction(&vm, "edit", std_edit)
	registerFunction(&vm, "print", std_print)
	registerFunction(&vm, "hold", std_hold)
	registerFunction(&vm, "unhold", std_unhold)
	registerFunction(&vm, "clear", std_clear)
	registerFunction(&vm, "rm", std_rm)

	var index int = eval(&vm, string(_txt))
	println(index)
}
