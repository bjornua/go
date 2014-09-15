package main

import (
    "io"
    "os"
    "strings"
    "bufio"
    "sort"
    "fmt"
    "sync"
    "flag"
    "runtime"
    "runtime/pprof"
)

//import _     "net/http/pprof"

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
    sorted   []rune
    original string
}

func NewSortedString(s string) sortedstring {
    var letters RuneSlice
    for _, x := range s {
        letters = append(letters, x)
    }
    sort.Sort(letters)
    return sortedstring{[]rune(letters), s}
}


func anagrammer(original string, words []string, maxdepth int) chan []string {
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

    r := make(chan []string, 10)
    lock := make(chan struct{}, 4)
    lock <- struct{}{}
    lock <- struct{}{}
    lock <- struct{}{}
    lock <- struct{}{}

    var wg sync.WaitGroup

    wg.Add(1)
    go anagrammer_r(lock, &wg, maxdepth, []string{}, r, charpool, xs_final)
    go func() {
        wg.Wait()
        close(r)
    }()
    return r
}

func haverunes(base []rune, subtrahend []rune) (bool) {
    if len(subtrahend) > len(base) {
        return false
    }
    j := 0
    for _, a := range subtrahend {
        for {
            if j == len(base) {
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
func removerunes(base []rune, subtrahend []rune) []rune{
    rest := make([]rune, 0, len(base))
    i := 0
    for _, a := range base {
        if i < len(subtrahend) && a == subtrahend[i] {
            i++
        } else {
            rest = append(rest, a)
        }
    }
    return rest
}
func filterwords(pool []rune, words []sortedstring) []sortedstring{
    validwords := make([]sortedstring, 0, len(words))
    for _, w := range words {
        if haverunes(pool, w.sorted) {
            validwords = append(validwords, w)
        }
    }
    return validwords
}

func anagrammer_r(lock chan struct{}, wg *sync.WaitGroup, maxdepth int, prefix []string, r chan []string, pool []rune, words []sortedstring) {
    validwords := filterwords(pool, words)
    for _, w := range validwords {
        new_prefix := append(prefix, w.original)
        newpool := removerunes(pool, w.sorted)
        if len(newpool) == 0 {
            r <- new_prefix
        } else if len(prefix) != maxdepth {
            wg.Add(1)
            go anagrammer_r(lock, wg, maxdepth, new_prefix, r, newpool, validwords)
            <- lock
        }
    }
    wg.Done()
    lock <- struct{}{}
}

func main() {
    runtime.GOMAXPROCS(4)

    f, err := os.Create("profile.pprof")
    if err != nil {
         panic(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    depth := flag.Int("depth", -1, "Maximum recursion depth")
    flag.Parse()
    anagram := flag.Arg(0)
    println(anagram)
    println(*depth)

    words := parseWordlist(os.Stdin)
    result := anagrammer(anagram, words, *depth)
    
    for r := range result {
        fmt.Print(r[0])
        for i := 1; i < len(r); i++ {
            fmt.Print(" ")
            fmt.Print(r[i])
       }
       fmt.Println()

    }

}