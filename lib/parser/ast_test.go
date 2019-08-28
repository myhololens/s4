package parser

import (
	"reflect"
	"testing"
)

func TestGenerateAST(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    []Token
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				input: "RUN 192.168.0.1",
			},
			want: []Token{
				{
					Key:   "RUN",
					value: []string{"192.168.0.1"},
				},
			},
		},
		{
			name: "basic with multiple blank",
			args: args{
				input: "RUN     192.168.0.1",
			},
			want: []Token{
				{
					Key:   "RUN",
					value: []string{"192.168.0.1"},
				},
			},
		},
		{
			name: "basic with comment single line",
			args: args{
				input: `# server host address
RUN     192.168.0.1`,
			},
			want: []Token{
				{
					Key:   "RUN",
					value: []string{"192.168.0.1"},
				},
			},
		},
		{
			name: "basic with tail comment",
			args: args{
				input: `RUN 192.168.0.1 # set your server address`,
			},
			want: []Token{
				{
					Key:   "RUN",
					value: []string{"192.168.0.1"},
				},
			},
		},
		{
			name: "basic with prefix comment",
			args: args{
				input: `# HOST 192.168.0.1`,
			},
			want: []Token{
			},
		},
		{
			name: "multiple field",
			args: args{
				input: `CONNECT axetroy@192.168.0.1:22
RUN ls -lh

`,
			},
			want: []Token{
				{
					Key:   "CONNECT",
					value: []string{"axetroy@192.168.0.1:22"},
				},
				{
					Key:   "RUN",
					value: []string{"ls", "-lh"},
				},
			},
		},
		{
			name: "multiple values",
			args: args{
				input: `UPLOAD ./README.md ./start.py ./dist`,
			},
			want: []Token{
				{
					Key:   "UPLOAD",
					value: []string{"./README.md", "./start.py", "./dist"},
				},
			},
		},
		{
			name: "invalid keyword",
			args: args{
				input: "INVALID KEYWORD",
			},
			want: []Token{
			},
			wantErr: true,
		},
		{
			name: "invalid keyword with comment",
			args: args{
				input: "# INVALID KEYWORD",
			},
			want: []Token{
			},
			wantErr: false,
		},
		{
			name: "Invalid grammar",
			args: args{
				input: "(abc)",
			},
			want: []Token{
			},
			wantErr: true,
		},
		{
			name: "Empty value",
			args: args{
				input: "CONNECT",
			},
			want: []Token{
			},
			wantErr: true,
		},
		{
			name: "Invalid ENV",
			args: args{
				input: "ENV PRIVATE_KEY",
			},
			want: []Token{
			},
			wantErr: true,
		},
		{
			name: "Invalid ENV",
			args: args{
				input: "ENV PRIVATE_KEY = ",
			},
			want: []Token{
			},
			wantErr: true,
		},
		{
			name: "parse ENV",
			args: args{
				input: "ENV PRIVATE_KEY = xxx",
			},
			want: []Token{
				{
					Key:   "ENV",
					value: []string{"PRIVATE_KEY", "xxx"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateAST(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAST() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateAST() = %v, want %v", got, tt.want)
			}
		})
	}
}