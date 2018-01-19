package bytepair

// Dictionary is the dictionary
type Dictionary []Entry

type Entry struct {
	Key   int32
	Value int32
}

func Encode(in []byte) (out []byte, dict Dictionary) {
	if len(in) < 4 {
		return in, dict
	}

	ints := stringToInt32s(in)

	var (
		// used is basically the count of unigrams so we can see what bytes are unused
		used [256]int32

		// p represents the previous bigram, the idea is that if two consecutive bigrams
		// are the same then it's likely 3 of the same bytes, and we don't want to count
		// the second one...e.g., aaa should just count the first aa as a byte pair
		p int32 = -1
	)

	// Find the bytes that are used, so we can use the unused bytes for encoding
	for _, u := range ints {
		used[u]++
	}

	// https://en.wikipedia.org/wiki/Byte_pair_encoding
	// Forever loop to go through the slice and identify the byte pair that occurs the
	// most often, encode that, then find the next set of byte pairs, etc
	for {
		// maxCnt is the number of occurrences for the byte pair that's occurred most often
		// maxPair is the byte pair that occurred most often
		maxCnt, maxPair := int32(0), int32(0)

		// bp contains the count for each byte pair, key is byte pair, value is count
		bp := make(map[int32]int32, 100)

		// b1 is byte #1 and b2 is byte #2 to calculate bigram
		b1 := ints[0]

		for _, b2 := range ints[1:] {
			// -1 means we have removed this byte during a loop, as in, this was the first byte
			// of a byte pair that got encoded into another byte. In this case, we will simply
			// skip it and since technically it's no longer part of the data
			if b2 == -1 {
				continue
			}

			// Calculate the bigram
			bigram := b1<<8 | b2

			// if the previous bigram is the same as the current bigram, that means we have 3
			// of the same bytes consecutively. In this case we cannot use the 2nd bigram because
			// the first byte of the bigram is already used as the second byte of the prvious bigram.
			if p == bigram {
				p = -1
				continue
			} else {
				p = bigram
			}

			b1 = b2
			bp[bigram]++

			// If the count for this bigram is more than others, then keep track of this one
			if n := bp[bigram]; n > 1 && n > maxCnt {
				maxCnt = n
				maxPair = bigram
			}
		}

		// if maxCnt is 0, that means we didn't find any recurring bigrams, so we are done.
		if maxCnt == 0 {
			break
		} else {
			// unused is a byte that does not occur in the data set
			unused := getUnused(used[:])

			// if unused == -1 that means we don't have any bytes we can use to represent, so we
			// have to end the encoding
			if unused == -1 {
				break
			}

			// If we have an unused byte we can use, then let's add that to the dictionary
			dict = append(dict, Entry{unused, maxPair})

			// For all the matching bigrams, we will set byte 1 to -1 (basically removing it),
			// and byte 2 to the unused byte
			for i, b1 := range ints[:len(ints)-1] {
				b2 := ints[i+1]
				b := b1<<8 | b2

				if b == maxPair {
					ints[i] = -1
					ints[i+1] = unused
					used[unused]++
				}
			}
		}
	}

	// Create the new buffer that contains the encoded output, i.e., removing all -1 bytes
	for _, u := range ints {
		if u != -1 {
			out = append(out, byte(u))
		}
	}

	return
}

func Decode(in []byte, dict Dictionary) (out []byte) {
	for i := len(dict) - 1; i >= 0; i-- {
		out = make([]byte, 0, len(in)*2)

		for _, b := range in {
			if int32(b) != dict[i].Key {
				out = append(out, b)
				continue
			}

			b1 := byte(dict[i].Value & 0xff)
			b2 := byte(dict[i].Value >> 8 & 0xff)

			out = append(out, b2, b1)
		}

		in = out
	}

	return in
}

// stringToInt32s converts a byte slice to an int32 slice
func stringToInt32s(blob []byte) (ints []int32) {
	for _, c := range blob {
		ints = append(ints, int32(c))
	}
	return
}

// getUnused find the first unused element (value == 0) in the slice of int32
func getUnused(ints []int32) int32 {
	for i, v := range ints {
		if v == 0 {
			return int32(i)
		}
	}

	return -1
}
