package jwd

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const variance = 0.001

func TestSort(t *testing.T) {

	Convey("sort should return shorter word in character count first", t, func() {

		Convey("it should work with words with onlyy ascii on them", func() {

			shorter, longer := sort("longerword", "short")

			So(longer, ShouldEqual, "longerword")
			So(shorter, ShouldEqual, "short")

		})

		Convey("it should work with words which have unicode in them", func() {

			shorter, longer := sort("ääni", "sound")

			So(longer, ShouldEqual, "sound")
			So(shorter, ShouldEqual, "ääni")

		})
	})
}

func TestWindow(t *testing.T) {

	Convey("testing window with precalculated data", t, func() {

		So(window("dicksonx"), ShouldEqual, float64(3))

	})
}

func TestClosestIndex(t *testing.T) {

	Convey("it should return the closest index", t, func() {

		t1 := closestIndex([]rune("eiiiiieii"), 'e', 4)
		So(t1, ShouldEqual, 6)

		t2 := closestIndex([]rune("dwayne"), 'e', 4)
		So(t2, ShouldEqual, 5)

		t3 := closestIndex([]rune("dwayne"), 'n', 3)
		So(t3, ShouldEqual, 4)

		t4 := closestIndex([]rune("sound"), 'n', 2)
		So(t4, ShouldEqual, 3)
	})

}

func TestScore(t *testing.T) {

	Convey("testing score with precalculated data", t, func() {

		m := float64(4)
		s1 := "dixon"
		s2 := "dicksonx"
		t := float64(0)

		So(score(m, t, float64(len(s1)), float64(len(s2))), ShouldAlmostEqual, 0.767, variance)

	})
}

func TestCalculate(t *testing.T) {

	Convey("Calculate should not return 0", t, func() {

		So(Calculate("dwayne", "duane"), ShouldAlmostEqual, 0.84, variance)
		So(Calculate("", "duane"), ShouldEqual, 0)
		So(Calculate("sound", "ääni"), ShouldAlmostEqual, 0.483, variance)
		So(Calculate("äiti", "ÄÄNI"), ShouldAlmostEqual, 0.75, variance)
		So(Calculate("äiti", "ääNI"), ShouldAlmostEqual, 0.75, variance)
		So(Calculate("dixon", "dicksonx"), ShouldAlmostEqual, 0.814, variance)
		So(Calculate("dicksonx", "dixon"), ShouldAlmostEqual, 0.814, variance)
		So(Calculate("DICKSONX", "DIXON"), ShouldAlmostEqual, 0.814, variance)
		So(Calculate("martha", "marhta"), ShouldAlmostEqual, 0.961, variance)
		So(Calculate("'foo", "fizz"), ShouldAlmostEqual, 0.166, variance)
		So(Calculate("jones", "johnson"), ShouldAlmostEqual, 0.832, variance)
		So(Calculate("asdfg", "qwerty"), ShouldEqual, 0)

	})
}

func ExampleCalculate() {
	distance := Calculate("jones", "johnson")
	fmt.Println(distance)
	// Output: 0.8323809523809523
}
