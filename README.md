# Velty (template engine in go)

[![GoReportCard](https://goreportcard.com/badge/github.com/viant/velty)](https://goreportcard.com/report/github.com/viant/velty)
[![GoDoc](https://godoc.org/github.com/viant/velty?status.svg)](https://godoc.org/github.com/viant/velty)

This library is compatible with Go 1.17+

Please refer to [`CHANGELOG.md`](CHANGELOG.md) if you encounter breaking changes.

- [Motivation](#motivation)
- [Introduction](#introduction)
- [Usage](#usage)
- [Performance](#benchmarks)
- [Tags](#tags)
- [Bugs](#bugs)
- [Contribution](#contributing-to-igo)

## Motivation

This library was created to facilitated seamless migration of code that uses JDK Velocity template to golang. 
The goal is to provide the first class template alternative for golang that is both substantially faster than JDK Velocity and
go standard template [HTML/Template](https://pkg.go.dev/html/template) or [Text/Template](https://pkg.go.dev/html/template)
See [benchmark](#benchmarks) section for details. 

## Introduction

In order to reduce execution time, this project first produces execution plan alongside with all variables needed to
execute it. One execution plan can be shared alongside many instances of scoped variables needed by executor. Scoped
Variables holds both execution state and variables defined or used in the evaluation code.

```go
    planner := velty.New()
    exec, newState, err := planner.Compile(code)
   
    state := newState() 
    exec.Exec(state)
    fmt.Printf("Result: %v", state.Buffer.String())
   
    anotherState := newState()
    exec.Exec(anotherState)
    fmt.Printf("Result: %v", anotherState.Buffer.String())
```

## Usage

In order to create execution plan, you need to create a planner:
```go
    planner := velty.New()
```

Options that you can pass while creating a Planner:
* `velty.BufferSize` - initial state buffer size
* `velty.CacheSize` - cache size for dynamically evaluated templates 
* `velty.EscapeHTML` - enables global (per Planner) HTML string escape mechanism (i.e. `$Foo`, if foo contains characters like `<>`, they will be encoded)

```go
    planner := velty.New(velty.BufferSize(1024), valty.CacheSize(200), velty.EscapeHTML(true))
```

Once you have the `Planner` you have to define variables that will be used. Velty doesn't use a `map` to store state, but it 
recreates an internal type each time you define new variable and uses `reflect.StructField.Offset` to access data from the state.
Velty supports two ways of defining planner variables:

* `planner.DefineVariable(variableName, variableType)` - will create and add non-anonymous `reflect.StructField`
* `planner.EmbedVariable(variableName, variableType)` - will create and add anonymous `reflect.StructField`

For each of the non-anonymous struct field registered with `DefineVariable` or `EmbedVariable` will be created unique `Selector`.
Selector is used to get field value from the state. 

```go
  err = planner.DefineVariable("foo", reflect.Typeof(Foo{})) 
  //handle error if needed
  err = planner.DefineVariable("boo", Boo{}) 
  //handle error if needed
  err = planner.EmbedVariable("bar", reflect.Typeof(Bar{})) 
  //handle error if needed
  err = planner.EmbedVariable("emp", Boo{}) 
  //handle error if needed
```

You can pass an instance or the `reflect.Type`. However, there are some constraints:
* Velty creates selector for each of the struct field. If you define i.e.:
```go
  type Foo struct {
      Name string
      ID int
  }
  
  planner.DefineVariable("foo", Foo{})
```
Velty will create three selectors: `foo`, `foo___Name`, `foo_ID`. Structure used by the velty shouldn't have three consecutive 
underscores in any of the fields.

* Velty won't create selectors for the Anonymous fields and will flatten the fields of the anonymous field.
```go
  type Foo struct {
      Name string
      ID int
  }
  
  type Bar struct {
      Foo
  }

  planner.EmbedVariable("foo", Bar{})
```
Velty will create only two selectors: `Name` and `ID` because all other fields are Anonymous.

* You can use tags to customize selector id, see: [Tags](#tags)
* Velty generates selectors for the constants and name them: `_T0`, `_T1`, `_T2` etc.

In the next step you can register functions. In the template you use the receiver syntax
i.e. `foo.Name.ToUpperCase()` but in the Go, you have to register plain function, where the first argument is the value of
field on which function was called.

```go
  err = planner.RegisterFunction("ToUpperCase", strings.ToUpper) 
  //handle error if needed
```

You can register function in two ways:
* `planner.RegisterFunction` - you can register regular functions like `strings.ToUpper`, and some of them are optimized using
type assertion. If the function isn't optimized, it will be called via `reflect.ValueOf.Call`. 

* `planner.RegsiterFunc` - if you notice that function is not optimized, you can optimize it registering `*op.Func`. 
The simple implementation:

```go
    customFunc := &op.Func{
		ResultType: reflect.TypeOf(""),
		Function: func(accumulator *Selector, operands []*Operand, state *est.State) unsafe.Pointer {
			if len(operands) < 2 {
				return nil
			}
			
			accumulator.SetBool(state.MemPtr, strings.HasPrefix(*(*string)(operands[0].Exec(state)), *(*string)(operands[1].Exec(state))))
			return accumulator.Pointer(state.MemPtr)
		},
	}
	
    err = planner.RegisterFunc("HasPrefix", customFunc) 
    //handle error if needed
```

Regular function can return no more than two non-pointer values. First is the new value, the second is an error. 
However errors in this case are ignored, and if any returned - the zero value will be appended to the result. 

The next step is to create execution plan and new state function:
```go
  template := `...`
  exec, newState, err := planner.Compile([]byte(template)) 
  // handle error if needed
  state := newState()
  exec.Exec(state)
```

## Tags
In order to match template identifiers with the struct fields, you can use the `velty` tag. 
Supported attributes:
* `name` - represents template identifier name i.e.:
```go
  type Foo struct {
    Name string `velty:"name=fooName"`
  }

  planner.DefineVariable("foo", Foo{})
  template := `${foo.fooName}`
```
* `names` - similar to the `name` but in this case you can specify more than one template identifier by separating them with `|`
```go
  type Foo struct {
    Name string `velty:"name=NAME|FOO_NAME"`
  }
   
  planner.DefineVariable("foo", Foo{})
  template := `${foo.NAME}, ${foo.FOO_NAME}`
```
* `prefix` - prefix can be used on the anonymous fields:
```go
  type Foo struct {
      Name string `velty:"name=NAME"`
  }
    
  type Boo struct {
      Foo `velty:"prefix=FOO_"`
  }

  planner.EmbedVariable("boo", Boo{})
  template := `${FOO_NAME}`
```

* `-` - tells Velty to don't create a selector for given field. In other words, it won't be possible to use the field in the template:
```go
 type Foo struct {
      Name string `velty:"-"`
  }

  planner.EmbedVariable("foo", Foo{})
  template := `${foo.Name}` // throws an error during compile time
```

## Benchmarks
Benchmarks against the `text/template` and `Java velocity`:


Bench 1: [The template](internal/bench/template/template.vm).

```
Benchmark_Exec_Velty-8   	   54585	         21127 ns/op	       0 B/op	       0 allocs/op	       4 allocs/op
Benchmark_Exec_Template-8   	    2370	        486511 ns/op	   78402 B/op	    3004 allocs/op
Benchmark_Exec_Velocity            44089                162599 ns/op
```


Bench 2: [The template](internal/bench/template/template_no_functions.vm).
```
Benchmark_Exec_Velty       	        69561	             16867 ns/op	       0 B/op	       0 allocs/op
Benchmark_Exec_Template   	        3103	            372839 ns/op	   66791 B/op	    2543 allocs/op
Benchmark_Exec_Velocity                62277                125636 ns/op
```


Bench 3: [The template](internal/bench/foreach/template.vm).

```
Benchmark_Exec_Velty     2077510       523.2 ns/op	       0 B/op	       0 allocs/op
Benchmark_Exec_Velocity  2077510       8183  ns/op
```

Velty template is substantially faster than JDK Velocity and go Text/Template.
On average velty is 20x faster than go [Text/template](https://pkg.go.dev/html/template)
and 8-15x faster than [JDK Apache Velocity](https://velocity.apache.org/)


## Optimizations

It will be possible to create pool of states (see  [Todo](.TODO.MD)) and reuse created states. 
This will reduce time needed to allocate new state.


## Bugs

This project does not implement full java velocity spec, but just a subset. It supports:
* variables - i.e. `${foo.Name} $Name`
* assignment - i.e. `#set($var1 = 10 + 20 * 10) #set($var2 = ${foo.Name})`
* if statements - i.e. `#if(1==1) abc #elsif(2==2) def #else ghi #end`
* foreach - i.e. `#foreach($name in ${foo.Names})`
* function calls - i.e. `${name.toUpper()}`
* template evaluation - i.e. `#evaluate($TEMPLATE)`

## Contributing to Velty

Velty is an open source project and contributors are welcome!

See [Todo](TODO.MD) list.


## License

The source code is made available under the terms of the Apache License, Version 2, as stated in the file `LICENSE`.

Individual files may be made available under their own specific license,
all compatible with Apache License, Version 2. Please see individual files for details.

## Credits and Acknowledgements

**Library Author:** Kamil Larysz


