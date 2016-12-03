package storage_provider

// a binary storage provider is used to store and
// retrieve binary data under a unique key
type Binary interface {
  // store "data" for "key"
  // should error if storage fails or if key is taken
  Set(key string, data []byte) error  
  
  // retrieve "data" for "key"
  // should error if retrieval fails
  Get(key string) ([]byte, error)
  
  // delete "data" for "key"
  // should error if deletion fails
  Unset(key string) error
  
  // checks if a key is in use
  Exists(key string) bool
}