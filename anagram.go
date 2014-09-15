package main

import (
    "io"
    "os"
    "strings"
    "bufio"
    "sort"
//    "fmt"
    "sync"
    "flag"
//    "bytes"
    "runtime/pprof"
)


func parseWordlist(r io.Reader) []string {
    var xs []string
    s := bufio.NewScanner(r)
    for s.Scan() {
        xs = append(xs, s.Text())
    }
    return xs
}


type RuneSlice []rune
func (p RuneSlice) Len() int           { return len(p) }
func (p RuneSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p RuneSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type sortedstring struct {
    sorted   RuneSlice
    original []byte
}

func NewSortedString(s string) sortedstring {
    var letters RuneSlice
    for _, x := range s {
        letters = append(letters, x)
    }
    sort.Sort(letters)
    return sortedstring{letters, []byte(s)}
}

func anagrammer(original string, words []string, maxdepth int) chan [][]byte {
    var xs []string
    for _, t := range words {
        t = strings.Trim(t, " \n")
        t = strings.ToLower(t)
        if len(t) > 0 {
            xs = append(xs, t)
        }
    }
    sort.Strings(xs)
    var xs_final []sortedstring
    for i := range xs {
        if i == 0 || xs[i] != xs[i-1] {
            xs_final = append(xs_final, NewSortedString(xs[i]))
        }
    }
    var charpool RuneSlice
    for _, c := range original {
        if c != ' ' {
            charpool = append(charpool, c)
        }
    }
    sort.Sort(charpool)
    r := make(chan [][]byte, 10)
    lock := make(chan struct{}, 4)
    mpool := make(chan *[]rune, 100)

    lock <- struct{}{}
    lock <- struct{}{}
    lock <- struct{}{}
    lock <- struct{}{}

    var wg sync.WaitGroup

    wg.Add(1)
    go anagrammer_r(mpool, lock, &wg, maxdepth, [][]byte{}, r, charpool, xs_final)
    go func() {
        wg.Wait()
        close(r)
    }()
    return r
}

func haverunes(base RuneSlice, subtrahend RuneSlice) (bool) {
    j := 0
    for _, a := range subtrahend {
        for {
            if j == len(base) || a < base[j] {
                return false
            }
            if a == base[j] {
                j++
                break
            }
            j++
        }
    }
    return true
}
func removerunes(base RuneSlice, subtrahend RuneSlice, rest *[]rune) {
    //*rest := make(RuneSlice, 0, len(base))
    i := 0
    for _, a := range base {
        if i < len(subtrahend) && a == subtrahend[i] {
            i++
        } else {
            *rest = append(*rest, a)
        }
    }
}

func getpool(pool chan *[]rune) *[]rune {
    select {
    case p := <- pool:
        *p = nil
        return p
    default:
        p := make([]rune, 0, 20)
        return &p
    }
}

func anagrammer_r(mpool chan *[]rune, lock chan struct{}, wg *sync.WaitGroup, maxdepth int, prefix [][]byte, r chan [][]byte, pool []rune, words []sortedstring) {
    validwords := make([]sortedstring, 0, len(words))
    for _, w := range words {
        if haverunes(pool, w.sorted) {
            validwords = append(validwords, w)
        }
    }
    for _, w := range validwords {
        new_prefix := append(prefix, w.original)
        newpool := *getpool(mpool)
        removerunes(pool, w.sorted, &newpool)
        if len(newpool) == 0 {
            r <- new_prefix
        } else if len(prefix) != maxdepth {
            wg.Add(1)
            go anagrammer_r(mpool, lock, wg, maxdepth, new_prefix, r, newpool, validwords)
            <- lock
        }
    }
    wg.Done()
    lock <- struct{}{}
    mpool <- &pool
}

func main() {
    f, err := os.Create("profile.pprof")
    if err != nil {
        panic(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    depth := flag.Int("depth", -1, "Maximum recursion depth")
    flag.Parse()
    anagram := flag.Arg(0)

    words := parseWordlist(os.Stdin)
    result := anagrammer(anagram, words, *depth)
    
    out := bufio.NewWriter(os.Stdout)
    space := []byte(" ")
    newline := []byte("\n")

    for r := range result {
        out.Write(r[0])
        for i := 1; i < len(r); i++ {
            out.Write(space)
            out.Write(r[i])
       }
       out.Write(newline)

        //fmt.Printf("%s = %s\n", string(anagram), x)
    }
    out.Flush()

    //fmt.Printf("%#v\n", words[:50])
}