package v8

import (
	"fmt"
	"sync"
	"testing"
)

var (
	JS_SIMPLE string = `a = 1; a++;`
	GO_FIB    string = `goFib(%d);`
	JS_FIB    string = `function fib(n) {
			var f = 0,
				n1 = 1,
				n2 = 0;
			if (n <= 1) {
				return n
			}

			for (var i = 1; i < n; i++) {
				f = n1 + n2;
				n2 = n1;
				n1 = f;

			}
			return f;
		}
		fib(%d);`
)

func TestEvalScript(t *testing.T) {
	ctx := NewContext()

	res, err := ctx.Eval(`var a = 10; a`)
	if err != nil {
		t.Fatal("Unexpected error on eval,", err)
	}
	if res == nil {
		t.Fatal("Expected result from eval, received nil")
	}

	switch res.(type) {
	case float64:
	default:
		t.Fatal("Expected float64 type")
	}
	if 10 != int(res.(float64)) {
		t.Fatal("Expected result to be 10, received:", res)
	}
}

func TestAddFunc(t *testing.T) {
	ctx := NewContext()

	err := ctx.AddFunc("_gov8_testFunc", func(args ...interface{}) interface{} {
		if len(args) != 2 {
			t.Fatal("Unexpected number of _gov8_testFunc's arguments.", len(args))
		}
		// First argument
		arg := args[0]
		switch arg.(type) {
		case float64:
		default:
			t.Fatal("Unexpected arg 0 type, expecting float64")
		}
		argVal := int(arg.(float64))
		if argVal != 10 {
			t.Fatal("Unexpected value for arg 0, expected 10, received:", argVal)
		}

		// Second argument
		arg = args[1]
		switch arg.(type) {
		case string:
		default:
			t.Fatal("Unexpected arg 1 type, expected string")
		}
		argVal2 := arg.(string)
		if argVal2 != "Test string" {
			t.Fatal("Unexpected value for arg 1, expected Test string, received:", argVal2)
		}

		return "testFunc return value"
	})
	if err != nil {
		t.Fatal("Expected to be able to add function, received error ", err)
	}

	res, err := ctx.Eval(`_gov8_testFunc(10, "Test string");`)
	if err != nil {
		t.Fatal("Unexpected error on testFunc eval,", err)
	}
	if res == nil {
		t.Fatal("Expected result from testFunc eval, received nil")
	}
	if res.(string) != "testFunc return value" {
		t.Fatal("Unexpected result from eval,", res)
	}
}

func TestAddFuncReturnArrayArgs(t *testing.T) {
	ctx := NewContext()

	err := ctx.AddFunc("testGoFunc", func(args ...interface{}) interface{} {
		return args
	})

	if err != nil {
		t.Fatal("Unexpected error when adding function", err)
	}

	res, err := ctx.Eval(`testGoFunc(10, "Test string");`)
	if err != nil {
		t.Fatal("Unexpected error when executing callback function from javascript", err)
	}
	if res == nil {
		t.Fatal("Expected result from callback function from javascript")
	}
	switch res.(type) {
	case []interface{}:
	default:
		t.Fatal("Unexpected type of arguments returned")
	}
	arrayRes := res.([]interface{})
	if len(arrayRes) != 2 {
		t.Fatal("Expected 2 items to be returned, received", len(arrayRes))
	}
	if arrayRes[0].(float64) != 10 {
		t.Fatal("Expected 10 as value for arrayRes[0], received", arrayRes[0])
	}
	if arrayRes[1].(string) != "Test string" {
		t.Fatal("Expected Test string as value for arrayRes[1], received", arrayRes[1])
	}
}

func TestAddFuncReturnObject(t *testing.T) {
	ctx := NewContext()
	err := ctx.AddFunc("testFunc", func(args ...interface{}) interface{} {
		return map[string]interface{}{
			"arg0": int(args[0].(float64)),
			"arg1": args[1].(string),
		}
	})
	if err != nil {
		t.Fatal("Expected to be able to add function, received error ", err)
	}

	res, err := ctx.Eval(`testFunc(10, "something").arg0`)
	if err != nil {
		t.Fatal("Unexpected error on testFunc eval ", err)
	}
	if res == nil {
		t.Fatal("Expected result from testFunc eval, received nil")
	}
	if int(res.(float64)) != 10 {
		t.Fatal("Expected result to be 10, got", res)
	}

	res, err = ctx.Eval(`testFunc(10, "something")`)
	if err != nil {
		t.Fatal("Unexpected error on testFunc eval ", err)
	}
	if res == nil {
		t.Fatal("Expected result from testFunc eval, received nil")
	}
	resMap := res.(map[string]interface{})
	arg0 := int(resMap["arg0"].(float64))
	if arg0 != 10 {
		t.Fatal("Expected arg0 value to be 10 got ", arg0)
	}
	arg1 := resMap["arg1"].(string)
	if arg1 != "something" {
		t.Fatal("Expected arg1 value to be something got ", arg1)
	}
}

func v8EvalRoutine(i int, wg *sync.WaitGroup, t *testing.T) {
	defer wg.Done()

	ctx := NewContext()
	res, err := ctx.Eval(fmt.Sprintf(JS_FIB, i))
	if err != nil {
		t.Fatal("Failed to evaluate test fib function for index,", i, "error:", err)
	}
	if res == nil {
		t.Fatal("Unexpected nil for result of test fib function for index", i)
	}
	r := uint64(res.(float64))
	if !((i == 80 && r == 23416728348467684) ||
		(i == 50 && r == 12586269025) ||
		(i == 20 && r == 6765) ||
		(i == 60 && r == 1548008755920)) {
		t.Fatal("Failed to calculate correct fib for index", i, "received value,", r)
	}
}

func TestMultiEvalAndRoutines(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go v8EvalRoutine(80, &wg, t)

	wg.Add(1)
	go v8EvalRoutine(50, &wg, t)

	wg.Add(1)
	go v8EvalRoutine(20, &wg, t)

	wg.Add(1)
	go v8EvalRoutine(60, &wg, t)

	wg.Wait()
}

func goFib(n int) uint64 {
	var v, n1, n2 uint64
	n1 = 1
	if n <= 1 {
		return uint64(n)
	}
	for i := 1; i < n; i++ {
		v = n1 + n2
		n2, n1 = n1, v
	}
	return v
}

func TestMultiFuncCallbackAndRouties(t *testing.T) {
	ctx := NewContext()
	ctx.AddFunc("goFib", func(args ...interface{}) interface{} {
		return goFib(int(args[0].(float64)))
	})

	fibs := [...]int{80, 50, 20, 60}

	var wg sync.WaitGroup
	for i := 0; i < len(fibs); i++ {
		wg.Add(1)
		go func(fibIdx int) {
			defer wg.Done()

			res, err := ctx.Eval(fmt.Sprintf(GO_FIB, fibIdx))
			if err != nil {
				t.Fatal("Failed to evaluate test fib function for index,", fibIdx, "error:", err)
			}
			if res == nil {
				t.Fatal("Unexpected nil for result of test fib function for index", fibIdx)
			}
			r := uint64(res.(float64))
			if !((fibIdx == 80 && r == 23416728348467684) ||
				(fibIdx == 50 && r == 12586269025) ||
				(fibIdx == 20 && r == 6765) ||
				(fibIdx == 60 && r == 1548008755920)) {
				t.Fatal("Failed to calculate correct fib for index", fibIdx, "received value,", r)
			}

		}(fibs[i])
	}

	wg.Wait()
}

func BenchmarkCreateContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewContext()
	}
}

func BenchmarkEvalSimple(b *testing.B) {
	b.StopTimer()
	ctx := NewContext()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		ctx.Eval(JS_SIMPLE)
	}
}
