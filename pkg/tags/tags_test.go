package tags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsTag(t *testing.T) {
	type args struct {
		tags []string
		tag  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				tags: []string{"tag1", "tag2", "tag3"},
				tag:  "tag1",
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				tags: []string{"tag1", "tag2", "tag3"},
				tag:  "tag4",
			},
			want: false,
		},
		{
			name: "test3",
			args: args{
				tags: []string{},
				tag:  "tag1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := ContainsTag(tt.args.tags, tt.args.tag); got != tt.want {
			t.Errorf("%q. containsTags() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestMergeTags(t *testing.T) {
	type args struct {
		existingTags []string
		newTags      []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				existingTags: []string{"tag1", "tag2", "tag3"},
				newTags:      []string{"tag1", "tag2", "tag3"},
			},
			want: []string{"tag1", "tag2", "tag3"},
		},
		{
			name: "test2",
			args: args{
				existingTags: []string{"tag1", "tag2", "tag3"},
				newTags:      []string{"tag4", "tag5", "tag6"},
			},
			want: []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6"},
		},
		{
			name: "test3",
			args: args{
				existingTags: []string{},
				newTags:      []string{"tag4", "tag5", "tag6"},
			},
			want: []string{"tag4", "tag5", "tag6"},
		},
		{
			name: "test4",
			args: args{
				existingTags: []string{"tag4", "tag5", "tag6"},
				newTags:      []string{},
			},
			want: []string{"tag4", "tag5", "tag6"},
		},
		{
			name: "test5",
			args: args{
				existingTags: []string{},
				newTags:      []string{},
			},
			want: []string{},
		},
		{
			name: "test6",
			args: args{
				existingTags: []string{"tag1", "tag2", "tag3"},
				newTags:      []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6"},
			},
			want: []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6"},
		},
	}
	for _, tt := range tests {
		if got := MergeTags(tt.args.existingTags, tt.args.newTags); !assert.ElementsMatch(t, got, tt.want) {
			t.Errorf("%q. AppendUniqueTags() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
