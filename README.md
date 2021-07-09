# Prolang
## Interpreted Programming Language (Implemented With GO )
By Ahmed Sermani 

## Features

- A High-Level Language
- Data Types
- Expressions
- Statements
- Variables
- Control Flow
- Functions
- Classes


## Installation
```
go install github.com/Ahmed-Sermani/prolang@latest
```

#### Usage

```
// interacive
prolang
> 
// Run File
prolang /path/to/file.pl
```

## Arithmatic & Expressions
```
print 123;     // 123
print -1;      // -1
print 123.456; // 123.456
print -0.001;  // -0.001
print 123 + 456; // 579
print 4 - 3; // 1
print 1.2 - 1.2; // 0
print 5 * 3; // 15
print 12.34 * 0.3; // 3.702
print 12.34 / 12.34;  // 1
print 8 / 2;         // 4
print "1" / 1; // Runtime Error: Operand must be a number.[line n].

// * has higher precedence than +.
print 2 + 3 * 4; // 14
// * has higher precedence than -.
print 20 - 3 * 4; // 8
// / has higher precedence than +.
print 2 + 6 / 3; // 4
// / has higher precedence than -.
print 2 - 6 / 3; // 0
// < has higher precedence than ==.
print false == 2 < 1; // true
// > has higher precedence than ==.
print false == 1 > 2; // true
// <= has higher precedence than ==.
print false == 2 <= 1; // true
// >= has higher precedence than ==.
print false == 1 >= 2; // true

// grouping.
print (2 * (6 - (2 + 2))); // 4

print "str" + "ing"; // string
print true + nil; // Runtime Error: Operands must be two numbers or two strings[line n]

print 1 < 2;    // true
print 2 < 2;    // false
print 1 <= 2;    // true
print 2 <= 1;    // false
print 1 > 2;    // false
print 2 > 1;    // true
print 1 >= 2;    // false
print 2 >= 1;    // true
print 1 == 1; // true
print 1 == 2; // false

print "str" == "str"; // true
print "str" == "ing"; // false
print nil == false; // false
print false == 0; // false
print 0 == "0"; // false

print nil == nil; // true
print true == true; // true
print true == false; // false
print !true;     // false
print !false;    // true
print !!true;    // true
print !123;      // false
print !0;        // false
```

## Control Flow / Functions

```
// scoping
{
  let a = "global a";
  let b = "global b";
  let c = "global c";
  {
    let a = "outer a";
    let b = "outer b";
      {
        let a = "inner a";
        print a; // inner a  
        print b; // outer b 
        print c; // global c 
    }
    print a; // outer a
    print b; // outer b
    print c; // global c 
}
  print a; // global a
  print b; // global b
  print c; // global c
}
// if
if (true) print "good"; // good
if (false) print "bad";

// block body.
if (true) { print "block"; } // expect: block

// Assignment in if condition.
let a = false;
if (a = true) print a; // expect: true

// if/else/logical operators/ comparison operators/ booleans
{
    let a = 20;
    let b = 90;

    if ((a + b > 80) and (a > 10)) {
      print "a + b > 80; a > 10"; 
    } else if ((a > 80) or (b > 80)) {
      print "(a > 80) or (b > 80)";
    }
}

{
    // function / recursive / if 
    func fib(n) {
      if (n <= 1) return n;
      return fib(n - 2) + fib(n - 1);
    }

    // for loop
    for (let i = 0; i < 20; i = i + 1) {
        print fib(i); // 0 , 1 , 1 , 2 ..etc
    }

    for (let i = 0; i < 20; i = i + 1) {
      print i;
    }
}

  // while loop
{
      let i = 5;
      print "Start";
      while (i > 0) {
          print i;
          i = i - 1;
      }
      print "Finish";
}

// Functions
func say(first, last) {
  print first + " " + last + "!";
}
say("Ahmad", "Sermani");

func fac(n) {
    if (n <= 2) return n;
    return n * fac(n - 1);
}
print fac(3); // expect: 6

// With Return
func x_2(x) { return x * x; }
print x_2(7); // expect: 49

// Closures
func makeCounter() {
  let i = 0;
  func count() {
    i = i + 1;
    print i;
  }

  return count;
}

let counter = makeCounter();
counter(); // "1".
counter(); // "2".
counter(); // "3".

```
## Classes & Inhertance
```
class Prolang {
  printHello() {
    return "Hello, World";
  }
}

print Prolang; // Prints "<class Prolang>"

// chains
print Prolang().printHello(); // Prints "Hello, World"
let instance = Prolang();
print instance; // Prints "<instance of Prolang>".
print instance.printHello(); // Prints "Hello, World"

// Properties and This
class NamePrinter {
  name() {
    let last = "Sermani";
    print this.first + " " + last + "!";
  }
}

let printer = NamePrinter();
printer.first = "Ahmed";
printer.name(); // Ahmed Sermani!

// Initializer
class WithInit {
    init(arg) {
        this.arg = arg;
    }
    getArg() {
        return this.arg;
    }
}
let wi = WithInit("Prolang");

print wi; // expect: <instance of WithInit>
print wi.arg; // expect: Prolang
print wi.getArg(); // expect: Prolang


class Person {
    init(name, age, city) {
        this.name = name;
        this.age = age;
        this.city = city;
    }
    adult() {
        return this.age >= 18; 
    }

    teenager() {
        return this.age >= 13 and this.age <= 19;
    }
    
    child() {
        return this.age <= 12;
    }

    str() {
        let category = nil;

        if (this.adult()) {
            category = "an adult";
        } else if (this.teenager()) {
            category = "a teenager";
        } else if (this.child()) {
            category = "a child";
        }

        return "My name is " + this.name +  " I am " + category + " living in " + this.city;
    }
}

let p = Person("Ali", 24, "Syria, Idlip");
print p.str(); // !expect: My name is Ali I am an adult living in Syria, Idlip

// Inheritance and Super
class Area {
    init(name, population) {
        this.name = name;
        this.population = population;
    }
    name() {
        return this.name;
    }
    population() {
        return this.population;
    }
    format() {
        return this.name + " has " + this.population + " inhabitants";
    }
}

class City extends Area {
    format() {
        return "The city " + super.format();
    }
}

let cty = City("Riyadh", "7000284");

print cty.format(); // The city Riyadh has 7000284 inhabitants
```
## License
MIT
