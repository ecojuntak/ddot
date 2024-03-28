package resolver

type ConstantError string

func (e ConstantError) Error() string {
	return string(e)
}

const (
	DNSTypeNotImplemented = ConstantError("DNS record type not implemented")
)
