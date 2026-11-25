package migrate

import "testing"

func TestRenderGrants(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		role    string
		want    string
		wantErr bool
	}{
		{
			name: "replaces every placeholder with quoted role",
			sql:  `GRANT USAGE ON SCHEMA users TO "__app_role__"; GRANT SELECT ON users.profiles TO "__app_role__";`,
			role: "moments_app",
			want: `GRANT USAGE ON SCHEMA users TO "moments_app"; GRANT SELECT ON users.profiles TO "moments_app";`,
		},
		{
			name: "escapes quotes in role name",
			sql:  `GRANT USAGE ON SCHEMA users TO "__app_role__";`,
			role: `bad"; DROP SCHEMA users; --`,
			want: `GRANT USAGE ON SCHEMA users TO "bad""; DROP SCHEMA users; --";`,
		},
		{
			name:    "file without placeholder",
			sql:     `GRANT USAGE ON SCHEMA users TO someone;`,
			role:    "moments_app",
			wantErr: true,
		},
		{
			name:    "empty role",
			sql:     `GRANT USAGE ON SCHEMA users TO "__app_role__";`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderGrants(tt.sql, tt.role)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}
