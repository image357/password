set(GO_LICENSES_DIR "${CMAKE_BINARY_DIR}/go_licenses")

add_custom_target(
        go_licenses ALL
        DEPENDS
        "${GO_LICENSES_DIR}/github.com/image357/password/LICENSE"
)

add_custom_command(
        OUTPUT
        "${GO_LICENSES_DIR}/github.com/image357/password/LICENSE"
        COMMAND go mod tidy
        COMMAND go install "github.com/google/go-licenses@latest"
        COMMAND "${CMAKE_COMMAND}" -E rm -rf "${GO_LICENSES_DIR}"
        COMMAND go-licenses check "github.com/image357/password/cinterface" --disallowed_types=forbidden,restricted,unknown
        COMMAND go-licenses save "github.com/image357/password/cinterface" --save_path="${GO_LICENSES_DIR}"
        DEPENDS
        ${GO_FILES}
        go.mod
)
