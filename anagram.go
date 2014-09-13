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
    original string
}
type sortedstringlist []sortedstring
/*func (p sortedstringlist) Len() int           { return len(p) }
func (p sortedstringlist) Less(i, j int) bool { return p[i].original < p[j].original }
func (p sortedstringlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }*/


func NewSortedString(s string) sortedstring {
    var letters RuneSlice
    for _, x := range s {
        letters = append(letters, x)
    }
    sort.Sort(letters)
    return sortedstring{letters, s}
}

func subtractletters(base RuneSlice, subtrahend sortedstring) (bool, RuneSlice) {
    var rest RuneSlice
    //rest := make(RuneSlice, 0, len(base))
    i := 0
    j := 0
    if  len(subtrahend.sorted) > len(base)  {
        return false, RuneSlice{}
    }
    for {
        if i >= len(subtrahend.sorted) {
            for _, r := range base[j:] {
                rest = append(rest, r)
            }
            break
        }
        if j >= len(base) {
            return false, RuneSlice{}
        }
        if base[j] < subtrahend.sorted[i] {
            rest = append(rest, base[j])
        } else if base[j] == subtrahend.sorted[i] {
            i++
        } else {
            return false, RuneSlice{}
        }
        j++
    }
    return true, rest
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
    var xs_final sortedstringlist
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
    r := make(chan []string)
    var wg sync.WaitGroup

    wg.Add(1)
    go anagrammer_r(&wg, maxdepth, []string{}, r, charpool, xs_final)
    go func() {
        wg.Wait()
        close(r)
    }()
    return r
}

func anagrammer_r(wg *sync.WaitGroup, maxdepth int, prefix []string, r chan []string, pool []rune, words sortedstringlist) {
    var validwords sortedstringlist
    var pools []RuneSlice
    for _, w := range words {
        success, newpool := subtractletters(pool, w)
        if success {
            pools = append(pools, newpool)
            validwords = append(validwords, w)
        }
    }
    for i, w := range validwords {
        newpool := pools[i]
        new_prefix := append(prefix, w.original)
        if len(newpool) == 0 {
            r <- new_prefix
        } else if len(prefix) != maxdepth {
            wg.Add(1)
            anagrammer_r(wg, maxdepth, new_prefix, r, newpool, validwords)
        }
    }
    wg.Done()
}

func main() {
    f, err := os.Create("profile.pprof")
    if err != nil {
        panic(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    depth := flag.Int("maxwords", -1, "Maximum words")
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

        //fmt.Printf("%s = %s\n", string(anagram), x)
    }

    //fmt.Printf("%#v\n", words[:50])
}