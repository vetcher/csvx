package csvx

import "testing"

func FuzzUnmarshalClients(f *testing.F) {
	f.Add([]byte("client_id,client_name,client_age\n1,Jose,42\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		var out []Client
		_ = Unmarshal(data, &out) // must not panic
	})
}
