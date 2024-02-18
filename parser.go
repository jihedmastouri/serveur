package main

type Field struct {
	name string
	kind FieldType
}

type Entity struct {
	Name   string
	Count  int
	Fields map[string]any
}

type FieldType string

const (
	StringType    FieldType = "str,string"
	NumberType              = "num,number"
	BooleanType             = "bool,boolean"
	NameType                = "name"
	UsernameType            = "username"
	FullnameType            = "fullname"
	EmailType               = "email"
	DateType                = "date"
	UrlType                 = "url"
	IpType                  = "ip"
	UuidType                = "uuid"
	IdType                  = "id"
	AddressType             = "address,addr"
	PhoneType               = "phone"
	ParagraphType           = "paragraph,pg"
)
