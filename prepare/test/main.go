package main

import (
	"fmt"
)

func main() {
	digits := []int{0, 1, 0, 3, 12}
	fmt.Println(moveZeroes(digits))
}
func moveZeroes(nums []int) []int {
	var count int = 0
	if len(nums) == 1 {
		return nums
	}
	for i := 0; i < len(nums); i++ {
		if nums[i] == 0 {
			count++
		}
	}
	for i := 0; i < len(nums); i++ {
		if nums[i] == 0 {
			for j := i; j < len(nums)-1; j++ {
				nums[j] = nums[j+1]
			}
		}
	}
	for i := 1; i <= count; i++ {
		nums[len(nums)-i] = 0
	}
	return nums
}
