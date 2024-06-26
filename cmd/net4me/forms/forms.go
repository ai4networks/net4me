package forms

var forms = map[string]func(){}

func addForm(name string, f func()) {
	forms[name] = f
}

func Forms() map[string]func() {
	return forms
}

func Form(name string) func() {
	return forms[name]
}
