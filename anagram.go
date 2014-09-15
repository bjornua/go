package main

import (
    "io"
    "os"
    "strings"
    "bufio"
    "sort"
    "sync"
    //"fmt"
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
    original []byte
}

func NewSortedString(s string) sortedstring {
    var letters RuneSlice
    for _, x := range s {
        letters = append(letters, x)
    }
    sort.Sort(letters)
    return sortedstring{[]rune(letters), []byte(s)}
}

func spawner(limit int) (func(func()), *sync.WaitGroup) {
    var wg sync.WaitGroup
    lock := make(chan struct{}, limit)
    for x := 0; x < limit; x++ {
        lock <- struct{}{}
    }
    wrapper := func(f func()){
        wg.Add(1)
        go func(){
            f()
            lock <- struct{}{}
            wg.Done()
        }()
        <- lock
    }
    return wrapper, &wg
}



func anagrammer(original string, words []string, maxdepth int) chan prefix {
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

    r := make(chan prefix, 10)
    spa, wg := spawner(8)

    go func() {
        spa(func(){
            anagrammer_r(spa, 0, maxdepth, nil, r, charpool, xs_final)
        })
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
            if j == len(base) || base[j] > a  {
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
type prefix struct {
    parent *prefix
    text []byte
}
func anagrammer_r(spa func(func()), depth int, maxdepth int, p *prefix, r chan prefix, pool []rune, words []sortedstring) {
    validwords := filterwords(pool, words)
    for _, w := range validwords {
        newprefix := prefix{p, w.original}
        newpool := removerunes(pool, w.sorted)
        if len(newpool) == 0 {
            r <- newprefix
        } else if depth != maxdepth {
            spa(func(){
                anagrammer_r(spa, depth+1, maxdepth, &newprefix, r, newpool, validwords)
            })
        }
    }
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
    w := bufio.NewWriter(os.Stdout)
    for p := range result {
        for {
            w.Write(p.text)
            w.Write([]byte(" "))
            if p.parent == nil {
                break
            }
            p = *p.parent
        }
        w.Write([]byte("\n"))
    }
    w.Flush()

}