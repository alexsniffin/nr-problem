package main

import (
	"reflect"
	"strings"
	"testing"
)

func Test_insertSort(t *testing.T) {
	type args struct {
		wscArr []wordSequenceCount
		wsc    wordSequenceCount
	}
	tests := []struct {
		name string
		args args
		want []wordSequenceCount
	}{
		{
			"shouldInsertSortMiddleElement",
			args{
				wscArr: []wordSequenceCount{
					{
						count: 3,
					}, {
						count: 1,
					},
				},
				wsc: wordSequenceCount{count: 2},
			},
			[]wordSequenceCount{
				{
					count: 3,
				}, {
					count: 2,
				}, {
					count: 1,
				},
			},
		},
		{
			"shouldInsertSortBeginning",
			args{
				wscArr: []wordSequenceCount{
					{
						count: 3,
					}, {
						count: 2,
					},
				},
				wsc: wordSequenceCount{count: 1},
			},
			[]wordSequenceCount{
				{
					count: 3,
				}, {
					count: 2,
				}, {
					count: 1,
				},
			},
		},
		{
			"shouldInsertSortEnd",
			args{
				wscArr: []wordSequenceCount{
					{
						count: 2,
					}, {
						count: 1,
					},
				},
				wsc: wordSequenceCount{count: 3},
			},
			[]wordSequenceCount{
				{
					count: 3,
				}, {
					count: 2,
				}, {
					count: 1,
				},
			},
		},
		{
			"shouldInsertSortOnly",
			args{
				wscArr: []wordSequenceCount{},
				wsc: wordSequenceCount{count: 1},
			},
			[]wordSequenceCount{
				{
					count: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := insertDescSort(tt.args.wscArr, tt.args.wsc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("insertDescSort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filter(t *testing.T) {
	t.Run("shouldRemoveSymbolEndOfWord", func(t *testing.T) {
		pattern := `[^\w]`
		c := make(chan []string)
		go func() {
			c <- []string{"a."}
			close(c)
		}()

		result := filter(pattern, c)

		arr := <-result
		if strings.Join(arr, " ") != "a" {
			t.Errorf("unexpected result, got %s, expected a", arr[0])
		}
	})
	t.Run("shouldRemoveSymbolMiddleOfWord", func(t *testing.T) {
		pattern := `[^\w]`
		c := make(chan []string)
		go func() {
			c <- []string{"a.b"}
			close(c)
		}()

		result := filter(pattern, c)

		arr := <-result
		if strings.Join(arr, " ") != "ab" {
			t.Errorf("unexpected result, got %s, expected ab", arr[0])
		}
	})
	t.Run("shouldRemoveSymbolManyWords", func(t *testing.T) {
		pattern := `[^\w]`
		c := make(chan []string)
		go func() {
			c <- []string{"a.", "b.", "c."}
			close(c)
		}()

		result := filter(pattern, c)

		arr := <-result
		if strings.Join(arr, " ") != "a b c" {
			t.Errorf("unexpected result, got %s, expected a b c", arr[0])
		}
	})
}

func Test_count(t *testing.T) {
	t.Run("shouldCountSingleWordSequence", func(t *testing.T) {
		c := make(chan []string)
		go func() {
			c <- []string{"a", "b", "c"}
			close(c)
		}()

		result := count(c)

		wsc := <-result
		if wsc.words != "a b c" || wsc.count != 1 {
			t.Errorf("unexpected result, got %v, expected wsc{a b c,1}", wsc)
		}
	})
	t.Run("shouldCountMultipleWordSequence", func(t *testing.T) {
		c := make(chan []string)
		go func() {
			c <- []string{"a", "b", "c", "d"}
			close(c)
		}()

		result := count(c)

		wsc1 := <-result
		wsc2 := <-result
		if wsc1.words != "a b c" || wsc1.count != 1 {
			t.Errorf("unexpected result, got %v, expected wsc{a b c,1}", wsc1)
		}
		if wsc2.words != "b c d" || wsc2.count != 1 {
			t.Errorf("unexpected result, got %v, expected wsc{b c d,1}", wsc2)
		}
	})
	t.Run("shouldCountMultipleOccurrenceWordSequence", func(t *testing.T) {
		c := make(chan []string)
		go func() {
			c <- []string{"a", "a", "a", "a"}
			close(c)
		}()

		result := count(c)

		wsc := <-result
		if wsc.words != "a a a" || wsc.count != 2 {
			t.Errorf("unexpected result, got %v, expected wsc{a a a,2}", wsc)
		}
	})
	t.Run("shouldCountMultipleWordSequenceManyArray", func(t *testing.T) {
		c := make(chan []string)
		go func() {
			c <- []string{"a", "b", "c"}
			c <- []string{"a", "b", "c"}
			close(c)
		}()

		result := count(c)

		wsc := <-result
		if wsc.words != "a b c" || wsc.count != 2 {
			t.Errorf("unexpected result, got %v, expected wsc{a b c,2}", wsc)
		}
	})
}

func Test_topN(t *testing.T) {
	t.Run("shouldGetSingleResult", func(t *testing.T) {
		c := make(chan wordSequenceCount)
		go func() {
			c <- wordSequenceCount{"a b c", 1}
			close(c)
		}()

		result := topN(1, c)

		if result[0].words != "a b c" || result[0].count != 1 {
			t.Errorf("unexpected result, got %v, expected wsc{a b c,1}", result[0])
		}
	})
	t.Run("shouldGetSingleResultDropOne", func(t *testing.T) {
		c := make(chan wordSequenceCount)
		go func() {
			c <- wordSequenceCount{"a b c", 2}
			c <- wordSequenceCount{"c b a", 1}
			close(c)
		}()

		result := topN(1, c)

		if result[0].words != "a b c" || result[0].count != 2 {
			t.Errorf("unexpected result, got %v, expected wsc{a b c,1}", result[0])
		}
	})
}