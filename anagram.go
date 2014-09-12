package main

import (
    "io"
    "os"
    "strings"
    "bufio"
    "sort"
    "fmt"
//    "bytes"
    "runtime/pprof"
)


func parseWordlist(r io.Reader) sortedstringlist  {
    var xs sortedstringlist
    s := bufio.NewScanner(r)
    for s.Scan() {
        t := s.Text()
        t = strings.Trim(t, " \n")
        t = strings.ToLower(t)
        xs = append(xs, NewSortedString(t))
    }
    return xs
}

type sortedstring struct {
    sorted   []rune
    original string
}
type sortedstringlist []sortedstring
func (p sortedstringlist) Len() int           { return len(p) }
func (p sortedstringlist) Less(i, j int) bool { return p[i].original < p[j].original }
func (p sortedstringlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }


type RuneSlice []rune
func (p RuneSlice) Len() int           { return len(p) }
func (p RuneSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p RuneSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }


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
    i := 0
    j := 0
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

func anagrammer(depth int, r chan string, pool []rune, words sortedstringlist) {
    var validwords []sortedstring
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
        rr := make(chan string, 2)
        if len(newpool) == 0 {
            r <- w.original
        } else if depth != 2 {
            go anagrammer(depth + 1, rr, newpool, validwords)
            for a := range rr {
                r <- w.original + " " + a
            }
        }
    }
    close(r)
}

func main() {
    f, err := os.Create("profile.pprof")
    if err != nil {
        panic(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    if len(os.Args) != 2 {
        println("Usage: ./rabbit-hole2 \"anagram\" < path/to/wordlist")
        os.Exit(2)
    }
    anagram := os.Args[1]

    var charpool RuneSlice
    for _, c := range anagram {
        if c != ' ' {
            charpool = append(charpool, c)
        }
    }
    sort.Sort(charpool)

    words := parseWordlist(os.Stdin)
    sort.Sort(words)

    result := make(chan string, 100)
    go anagrammer(0, result, charpool, words)
    
    for x := range result {
        fmt.Printf("%s = %s\n", string(anagram), x)
    }

    //fmt.Printf("%#v\n", words[:50])
}