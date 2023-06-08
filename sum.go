package nex

func sum(slice []byte) int {
	total := 0
	for _, value := range slice {
		total += int(value)
	}

	return total
}
