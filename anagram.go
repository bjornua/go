package main

import (
    "io"
    "os"
    "strings"
    "bufio"
    "sort"
    "fmt"
    "bytes"
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
    sorted   string
    original string
}
type sortedstringlist []sortedstring
func (p sortedstringlist) Len() int           { return len(p) }
func (p sortedstringlist) Less(i, j int) bool { return p[i].original < p[j].original }
func (p sortedstringlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func NewSortedString(s string) sortedstring {
    var letters []string
    for _, x := range s {
        letters = append(letters, string(x))
    }
    sort.Strings(letters)
    return sortedstring{strings.Join(letters, ""), s}
}

func subtractletters(base string, subtrahend sortedstring) (bool, string) {
    var rest bytes.Buffer
    i := 0
    j := 0
    for {
        if i >= len(subtrahend.sorted) {
            rest.WriteString(base[j:])
            break
        }
        if j >= len(base) {
            return false, ""
        }
        
        b := base[j]
        s := subtrahend.sorted[i]

        if b < s {
            rest.WriteString(string(b))
            j++
        } else if b > s {
            return false, ""
        } else {
            j++
            i++
        }
    }
    return true, rest.String()
}

type anagram struct {
    word sortedstring
    pool string
}
func anagrammer(depth int, r chan string, pool string, words sortedstringlist) {
    var validwords []sortedstring
    var pools []string
    for _, w := range words {
        success, newpool := subtractletters(pool, w)
        if !success {
            continue;
        }
        validwords = append(validwords, w)
        pools = append(pools, newpool)
    }
    for i, w := range validwords {
        newpool := pools[i]
        rr := make(chan string)
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

    words := parseWordlist(os.Stdin)
    sort.Sort(words)

    result := make(chan string)
    go anagrammer(0, result, strings.Trim(NewSortedString(anagram).sorted, " "), words)
    
    for x := range result {
        fmt.Printf("%s = %s\n", anagram, x)
    }

    //fmt.Printf("%#v\n", words[:50])
}