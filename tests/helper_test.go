package testhelpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateConfigFile(t *testing.T) {
	type args struct {
		config     any
		configFile string
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "test json content",
			args: args{
				configFile: "config1.json",
				config: map[string]any{
					"string": "config_value",
					"int":    float64(10),
					"float":  10.1,
					"bool":   true,
				},
			},
			wantErr: false,
		},
		{
			name: "test raw content, byte",
			args: args{
				configFile: "config2.json",
				config: []byte(`{
					"string": "config_value",
					"int":    10,
					"float":  10.10,
					"bool":   true
				}`),
			},
			wantErr: false,
		},
		{
			name: "test raw content string",
			args: args{
				configFile: "config3.json",
				config: `{
					"string": "config_value",
					"int":    10,
					"float":  10.10,
					"bool":   true
				}`,
			},
			wantErr: false,
		},
		{
			name: "test bad file name",
			args: args{
				configFile: os.Args[0],
				config: `{
					"string": "config_value",
					"int":    10,
					"float":  10.10,
					"bool":   true
				}`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := filepath.Join(t.TempDir(), tt.args.configFile)
			if err := CreateConfigFile(configFile, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("CreateConfigFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				confFile, err := os.Open(configFile)
				require.NoError(t, err)
				defer func() {
					err = confFile.Close()
					require.NoError(t, err)
				}()
				chekCfg := make(map[string]any)
				require.NoError(t, json.NewDecoder(confFile).Decode(&chekCfg))

				whantCfg := make(map[string]any)

				switch c := tt.args.config.(type) {
				case []byte:
					err = json.Unmarshal(c, &whantCfg)
				case string:
					err = json.Unmarshal([]byte(c), &whantCfg)
				case map[string]any:
					whantCfg = c
				default:
					b, er := json.Marshal(c)
					require.NoError(t, er)
					err = json.Unmarshal(b, &whantCfg)
				}
				require.NoError(t, err)
				assert.Equal(t, true,
					fmt.Sprintf("%v", chekCfg) == fmt.Sprintf("%v", whantCfg),
					reflect.DeepEqual(chekCfg, whantCfg),
					fmt.Sprintf("expected: %v\n  actual: %v", chekCfg, whantCfg))
			}
		})
	}
}
