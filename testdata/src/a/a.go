package a

type person struct {
	name string
	age  int
}

func f() bool {
	return true
}

func ff() (bool, bool, error) {
	return true, true, nil
}
