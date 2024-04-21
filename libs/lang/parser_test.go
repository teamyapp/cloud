package lang

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser_Parse(t *testing.T) {
	testCases := []struct {
		name               string
		source             string
		expectedStatements []string
		hasError           bool
	}{
		{
			name:               "simple expression",
			source:             `2 + 5 * 6;`,
			expectedStatements: []string{"(+ 2 (* 5 6))"},
		},
		{
			name:               "simple expression with grouping",
			source:             `(2 + 5) * 6;`,
			expectedStatements: []string{"(* (group (+ 2 5)) 6)"},
		},
		{
			name:               "simple expression with unary",
			source:             `!!true;`,
			expectedStatements: []string{"(! (! true))"},
		},
		{
			name:               "simple equality expression",
			source:             `2 == 5;`,
			expectedStatements: []string{"(== 2 5)"},
		},
		{
			name:               "simple comparison expression",
			source:             `2 > 5;`,
			expectedStatements: []string{"(> 2 5)"},
		},
		{
			name:               "simple equality expression with comparison",
			source:             `2 >= 5 == true;`,
			expectedStatements: []string{"(== (>= 2 5) true)"},
		},
		{
			name:               "simple comparison expression with term",
			source:             `2 + 5 > 6;`,
			expectedStatements: []string{"(> (+ 2 5) 6)"},
		},
		{
			name:               "simple term expression with factor",
			source:             `2 * 5 + 6;`,
			expectedStatements: []string{"(+ (* 2 5) 6)"},
		},
		{
			name:               "simple factor expression with unary",
			source:             `2 * -5;`,
			expectedStatements: []string{"(* 2 (- 5))"},
		},
		{
			name:               "simple string equality expression",
			source:             `"Hello World!" == "Hello" + " " + "World!";`,
			expectedStatements: []string{`(== "Hello World!" (+ (+ "Hello" " ") "World!"))`},
		},
		{
			name:               "complex expression",
			source:             `!(2 * (5 + 6) <= 4) == true;`,
			expectedStatements: []string{"(== (! (group (<= (* 2 (group (+ 5 6))) 4))) true)"},
		},
		{
			name:     "expression with unmatched (",
			source:   `!(2 * (5 + 6 <= 4) == true;`,
			hasError: true,
		},
		{
			name:               "expressions separated by comma",
			source:             `2 + 5, 6 * 7, 8 == 9, !true, "Hello";`,
			expectedStatements: []string{`(expressionList (+ 2 5) (* 6 7) (== 8 9) (! true) "Hello")`},
		},
		{
			name:               "simple ternary expression",
			source:             `true ? 2 : 3;`,
			expectedStatements: []string{"(?: true 2 3)"},
		},
		{
			name:               "complex ternary expression",
			source:             `1 == 2 ? 2 + 5 * 6 : 3 + 4 / 2;`,
			expectedStatements: []string{"(?: (== 1 2) (+ 2 (* 5 6)) (+ 3 (/ 4 2)))"},
		},
		{
			name:               "chained ternary expression",
			source:             `1 == 2 ? 2 + 5 * 6 : 3 + 4 / 2 ? 1 : 2;`,
			expectedStatements: []string{"(?: (== 1 2) (+ 2 (* 5 6)) (?: (+ 3 (/ 4 2)) 1 2))"},
		},
		{
			name:               "chained ternary expression with group",
			source:             `1 == 2 ? 2 + 5 * 6 : (3 + 4 / 2 ? 1 : 2);`,
			expectedStatements: []string{"(?: (== 1 2) (+ 2 (* 5 6)) (group (?: (+ 3 (/ 4 2)) 1 2)))"},
		},
		{
			name:     "ternary expression with unmatched :",
			source:   "1 == 2 ? 2 + 5 * 6;",
			hasError: true,
		},
		{
			name:     "add without left operand",
			source:   `+ 5;`,
			hasError: true,
		},
		{
			name:     "add without right operand",
			source:   `5 +;`,
			hasError: true,
		},
		{
			name:     "unary without operand",
			source:   `!;`,
			hasError: true,
		},
		{
			name:               "simple && and ||",
			source:             `true || false && true;`,
			expectedStatements: []string{"(|| true (&& false true))"},
		},
		{
			name:               "simple &, | and ^",
			source:             `1 | 2 ^ 3 & 4;`,
			expectedStatements: []string{"(| 1 (^ 2 (& 3 4)))"},
		},
		{
			name:               "simple << and >>",
			source:             `1 << 2 >> 3;`,
			expectedStatements: []string{"(>> (<< 1 2) 3)"},
		},
		{
			name: "multiple statements",
			source: `
				print(2 + 5 * 6); 
				2 == 5;`,
			expectedStatements: []string{
				"(call print [(+ 2 (* 5 6))])",
				"(== 2 5)",
			},
		},
		{
			name: "define and read variables",
			source: `
				let a = 5;
				let b = 6;
				let c = a + b;
				print(c * 2);
`,
			expectedStatements: []string{
				"(let a 5)",
				"(let b 6)",
				"(let c (+ a b))",
				"(call print [(* c 2)])",
			},
		},
		{
			name: "define, assign value to and read from variables",
			source: `
				let a = 5;
				let b = 6;
				a = 1 + 2 * b;
				b = a - 5;
				print(a + " " + b);
`,
			expectedStatements: []string{
				"(let a 5)",
				"(let b 6)",
				"(= a (+ 1 (* 2 b)))",
				"(= b (- a 5))",
				`(call print [(+ (+ a " ") b)])`,
			},
		},
		{
			name: "variable scope with block statement",
			source: `
				let a = 5;
				let b = 6;
				{
					let a = 8;
					b = b + 1;
					print(a + " " + b);
				}
				print(" ");
				print(a + " " + b);
			`,
			expectedStatements: []string{
				"(let a 5)",
				"(let b 6)",
				`{(let a 8) (= b (+ b 1)) (call print [(+ (+ a " ") b)])}`,
				`(call print [" "])`,
				`(call print [(+ (+ a " ") b)])`,
			},
		},
		{
			name: "if without else",
			source: `
				if (1 + 2 > 0) {
					print("first true");

					if (3 * 4 > 1) {
						print(" | second true");
					}
				}
			`,
			expectedStatements: []string{
				`(if (> (+ 1 2) 0) {(call print ["first true"]) (if (> (* 3 4) 1) {(call print [" | second true"])})})`,
			},
		},
		{
			name: "if with else",
			source: `
				if (1 + 2 > 0) {
					print("first true");

					if (3 * 4 < 1) {
						print(" | second true");
					} else {
						print(" | second false");
					}
				} else {
					print("first false");
				}
			`,
			expectedStatements: []string{
				`(if (> (+ 1 2) 0) {(call print ["first true"]) (if (< (* 3 4) 1) {(call print [" | second true"])} else {(call print [" | second false"])})} else {(call print ["first false"])})`,
			},
		},
		{
			name: "while loop",
			source: `
				let a = 0;
				while (a < 5) {
					print(a);
					a = a + 1;
				}

				a;
`,
			expectedStatements: []string{
				"(let a 0)",
				`(while (< a 5) {(call print [a]) (= a (+ a 1))})`,
				"a",
			},
		},
		{
			name: "for loop",
			source: `
				for (let i = 0; i < 5; i = i + 1) {
					let num = i * 2;
					print(num);
				}
			`,
			expectedStatements: []string{
				`{(let i 0) (while (< i 5) {{(let num (* i 2)) (call print [num])} (= i (+ i 1))})}`,
			},
		},
		{
			name: "while loop with break",
			source: `
				let a = 0;
				while (true) {
					if (a >= 5) {
						break;
					}
					
					print(a);
					a = a + 1;
				}
			`,
			expectedStatements: []string{
				"(let a 0)",
				`(while true {(if (>= a 5) {(break)}) (call print [a]) (= a (+ a 1))})`,
			},
		},
		{
			name: "for loop with break",
			source: `
				for (let i = 0; true; i = i + 1) {
					if (i >= 5) {
						break;
					}
					
					let num = i * 2;
					print(num);
				}
			`,
			expectedStatements: []string{
				`{(let i 0) (while true {{(if (>= i 5) {(break)}) (let num (* i 2)) (call print [num])} (= i (+ i 1))})}`,
			},
		},
		{
			name: "call native function",
			source: `
				now();
			`,
			expectedStatements: []string{
				"(call now [])",
			},
		},
		{
			name: "define and call function",
			source: `
				hello()(1);
				add(5, 6);

				func hello() {
					print("Hello world!");
				}

				func add(a, b) {
					let c = 10;
					return a + b + c;
				}
			`,
			expectedStatements: []string{
				"(call (call hello []) [1])",
				"(call add [5 6])",
				`(func hello [] {(call print ["Hello world!"])})`,
				`(func add [a b] {(let c 10) (return (+ (+ a b) c))})`,
			},
		},
		{
			name: "define and call lambda",
			source: `
				func printNums(calc) {
		            for (let i = 0; i < 5; i = i + 1) {
		                print(calc(i));
		            }
				}

				printNums(func (num) {
					return num * 2;
				});
			`,
			expectedStatements: []string{
				`(func printNums [calc] {{(let i 0) (while (< i 5) {{(call print [(call calc [i])])} (= i (+ i 1))})}})`,
				`(call printNums [(lambda [num] {(return (* num 2))})])`,
			},
		},
		{
			name: "define lambda as expression statement",
			source: `
				func (num) {
					print(num * 2);
				}(10);
			`,
			expectedStatements: []string{
				`(call (lambda [num] {(call print [(* num 2)])}) [10])`,
			},
		},
		{
			name: "define class",
			source: `
				class Counter {
					constructor(initialCount) {
						this.count = initialCount;
						print(initialCount);
					}

					increment() {
						this.count = this.count + 1;
					}

					decrement() {
						this.count = this.count - 1;
					}

					getCount() {
						return this.count;
					}
				}
				
				let counter = new Counter(10);`,
			expectedStatements: []string{
				"(class Counter [instance (methods (func constructor [initialCount] {(set (this) count initialCount) (call print [initialCount])}) (func increment [] {(set (this) count (+ (get (this) count) 1))}) (func decrement [] {(set (this) count (- (get (this) count) 1))}) (func getCount [] {(return (get (this) count))})) (getters ) (setters )] [static (methods ) (getters ) (setters )])",
				"(let counter (new Counter [10]))",
			},
		},
		{
			name: "get object field",
			source: `
				class Counter {
					constructor(initialCount) {
						this.count = initialCount;
					}

					increment() {
						this.count = this.count + 1;
					}
				}

				class CounterBox {
					constructor(initialCounter) {
						this.counter = initialCounter;
					}
				}
				
				let counter = new Counter(10);
				let counterBox = new CounterBox(counter);
				counterBox.counter.count;
			`,
			expectedStatements: []string{
				"(class Counter [instance (methods (func constructor [initialCount] {(set (this) count initialCount)}) (func increment [] {(set (this) count (+ (get (this) count) 1))})) (getters ) (setters )] [static (methods ) (getters ) (setters )])",
				"(class CounterBox [instance (methods (func constructor [initialCounter] {(set (this) counter initialCounter)})) (getters ) (setters )] [static (methods ) (getters ) (setters )])",
				"(let counter (new Counter [10]))",
				"(let counterBox (new CounterBox [counter]))",
				"(get (get counterBox counter) count)",
			},
		},
		{
			name: "set object field",
			source: `
				class Counter {
					constructor(initialCount) {
						this.count = initialCount;
					}

					increment() {
						this.count = this.count + 1;
					}
				}

				class CounterBox {
					constructor(initialCounter) {
						this.counter = initialCounter;
					}
				}
				
				let counter = new Counter(10);
				let counterBox = new CounterBox(counter);
				counterBox.counter.count = 20;
			`,
			expectedStatements: []string{
				"(class Counter [instance (methods (func constructor [initialCount] {(set (this) count initialCount)}) (func increment [] {(set (this) count (+ (get (this) count) 1))})) (getters ) (setters )] [static (methods ) (getters ) (setters )])",
				"(class CounterBox [instance (methods (func constructor [initialCounter] {(set (this) counter initialCounter)})) (getters ) (setters )] [static (methods ) (getters ) (setters )])",
				"(let counter (new Counter [10]))",
				"(let counterBox (new CounterBox [counter]))",
				"(set (get counterBox counter) count 20)",
			},
		},
		{
			name: "define and call static method in class",
			source: `
				class Counter {
					constructor(initialCount) {
						this.count = initialCount;
					}

					increment() {
						this.count = this.count + 1;
					}

					static getClassName() {
						return "Counter";
					}
				}

				Counter.getClassName();
			`,
			expectedStatements: []string{
				"(class Counter [instance (methods (func constructor [initialCount] {(set (this) count initialCount)}) (func increment [] {(set (this) count (+ (get (this) count) 1))})) (getters ) (setters )] [static (methods (func getClassName [] {(return \"Counter\")})) (getters ) (setters )])",
				"(call (get Counter getClassName) [])",
			},
		},
		{
			name: "define and call getters and setters in class",
			source: `
				class Rectangle3D {
					constructor(width, height) {
						this.width = width;
						this.height = height;
					}

					get area {
						return this.width * this.height;
					}

					get volume {
						return this.width * this.height * this.depth;
					}

					set depth(value) {
						this.depth = value;
					}
				}

				let rectangle = new Rectangle3D(5, 6);
				rectangle.area;
				rectangle.depth = 10;
				rectangle.volume;
			`,
			expectedStatements: []string{
				"(class Rectangle3D [instance (methods (func constructor [width height] {(set (this) width width) (set (this) height height)})) (getters (func area [] {(return (* (get (this) width) (get (this) height)))}) (func volume [] {(return (* (* (get (this) width) (get (this) height)) (get (this) depth)))})) (setters (func depth [value] {(set (this) depth value)}))] [static (methods ) (getters ) (setters )])",
				"(let rectangle (new Rectangle3D [5 6]))",
				"(get rectangle area)",
				"(set rectangle depth 10)",
				"(get rectangle volume)",
			},
		},
		{
			name: "inherit from super class",
			source: `
				class Shape {
					getName() {
						return "Shape";
					}
				}

				class Rectangle : Shape {
					getName() {
						let name = super.getName();
						return name + " " + "Rectangle";
					}
				}

				let rectangle = new Rectangle();
				rectangle.getName();
			`,
			expectedStatements: []string{
				`(class Shape [instance (methods (func getName [] {(return "Shape")})) (getters ) (setters )] [static (methods ) (getters ) (setters )])`,
				`(class Rectangle [super Shape] [instance (methods (func getName [] {(let name (call (super getName) [])) (return (+ (+ name " ") "Rectangle"))})) (getters ) (setters )] [static (methods ) (getters ) (setters )])`,
				"(let rectangle (new Rectangle []))",
				"(call (get rectangle getName) [])",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tokens, errs := NewScanner().ScanTokens(tc.source)
			require.Equal(t, 0, len(errs))

			parser := NewParser()
			statements, errs := parser.Parse(tokens)
			if tc.hasError {
				require.True(t, len(errs) > 0)
				return
			}

			require.Equal(t, 0, len(errs))

			for index, statement := range statements {
				require.Equal(t, tc.expectedStatements[index], statement.String())
			}
		})
	}
}
