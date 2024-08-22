package adapter

type BeforePreStarter interface { //karing
	BeforePreStart() error
}

type PreStarter interface {
	PreStart() error
}

type PostStarter interface {
	PostStart() error
}
