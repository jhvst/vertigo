/*
Package jwd implements native [Jaro-Winkler distance](https://en.wikipedia.org/wiki/Jaro%E2%80%93Winkler_distance) in Go.

The original paper for Jaro-Winkler distance from 1990 can be found here: http://www.amstat.org/sections/srms/Proceedings/papers/1990_056.pdf

However, this implementation is rather written with the help of [this JavaScript library](https://github.com/NaturalNode/natural/blob/master/lib/natural/distance/jaro-winkler_distance.js) and the Wikipedia article mentioned above. Therefore, some differences might occur in distance scores.
*/

package jwd
