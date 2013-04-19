// Safety checks
package rtlib

// AliasCheck returns true if both arguments do not alias
func AliasCheck(a interface{}, b interface{}) bool {
	return true // TODO: actually check
}
