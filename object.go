package isolator

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var (
	DefaultObjectBuilder = NewClassicObjectBuilder()
)

type Object interface {
	Derive(session *Session) (obj Object, err error)
}

type ObjectBuilder interface {
	RegisterObjects(objects ...Object) (err error)
	DeriveObjects(session *Session, types ...reflect.Type) (objects []Object, err error)
}

type ClassicObjectBuilder struct {
	objects      map[string]Object
	objectLocker sync.Mutex
}

func NewClassicObjectBuilder() ObjectBuilder {
	return &ClassicObjectBuilder{
		objects: make(map[string]Object),
	}
}

func RegisterObjects(objects ...Object) (err error) {
	return DefaultObjectBuilder.RegisterObjects(objects...)
}

func (p *ClassicObjectBuilder) RegisterObjects(objects ...Object) (err error) {
	if objects == nil {
		return
	}

	p.objectLocker.Lock()
	defer p.objectLocker.Unlock()

	for i, object := range objects {
		if object == nil {
			err = errors.New("object is nil, index:" + strconv.Itoa(i))
			return
		}

		objectName := ""
		if objectName, err = getStructName(object); err != nil {
			return
		}

		objectName = strings.TrimSpace(objectName)

		if len(objectName) == 0 {
			err = errors.New("object name is empty")
			return
		}

		if _, exist := p.objects[objectName]; exist {
			err = errors.New("object" + objectName + "already exist")
			return
		}

		p.objects[objectName] = object
	}

	return
}

func (p *ClassicObjectBuilder) DeriveObjects(session *Session, types ...reflect.Type) (objects []Object, err error) {
	if types == nil {
		return
	}

	var objs []Object
	for _, typ := range types {

		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		var parentObj Object
		var exist bool

		if typ.Kind() == reflect.Interface {
			for _, obj := range p.objects {
				tpObj := reflect.TypeOf(obj)
				if tpObj.ConvertibleTo(typ) {
					parentObj = obj
					break
				}
			}
		} else if parentObj, exist = p.objects[typ.String()]; !exist {
			err = errors.New("object of " + typ.String() + " did not register")
			return
		}

		if parentObj == nil {
			err = errors.New("could not found useable object of " + typ.String())
			return
		}

		var childObj Object
		if childObj, err = parentObj.Derive(session); err != nil {
			return
		}
		objs = append(objs, childObj)
	}

	objects = objs
	return
}
