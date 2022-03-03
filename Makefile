copy:
		cp -r /usr/local/go/src/reflect ./internal/
		cp -r /usr/local/go/src/internal/abi ./internal/
		cp -r /usr/local/go/src/internal/goarch ./internal/
		cp -r /usr/local/go/src/internal/goexperiment ./internal/
		cp -r /usr/local/go/src/internal/unsafeheader ./internal/
		cp -r /usr/local/go/src/internal/bytealg ./internal/
		cp -r /usr/local/go/src/internal/itoa ./internal/
		cp -r /usr/local/go/src/internal/cpu ./internal/
		go-imports-rename -s --root ./internal 'internal/ => github.com/sirkon/deepequal/internal/'


# Need to extend deepValueEqual with `if pbms1, ok : …` after

#// Tests for deep equality using reflected types. The map argument tracks
#// comparisons that have already been seen, which allows short circuiting on
#// recursive types.
#func deepValueEqual(v1, v2 Value, visited map[visit]bool) bool {
#       if !v1.IsValid() || !v2.IsValid() {
#               return v1.IsValid() == v2.IsValid()
#       }
#       if v1.Type() != v2.Type() {
#               return false
#       }
#
#       if pbmsg1, ok := getProtoMessage(digValue(v1)); ok {
#               if pbmsg2, ok := getProtoMessage(digValue(v2)); ok {
#                       if pbmsg1 == nil {
#                               return pbmsg2 == nil
#                       } else if pbmsg2 == nil {
#                               return false
#                       }
#                       return proto.Equal(pbmsg1, pbmsg2)
#               }
#       }