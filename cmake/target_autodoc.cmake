add_custom_target(
        autodoc
        DEPENDS
        "${CMAKE_SOURCE_DIR}/docs/password.md"
        "${CMAKE_SOURCE_DIR}/docs/rest.md"
        "${CMAKE_SOURCE_DIR}/docs/log.md"
        "${CMAKE_SOURCE_DIR}/docs/cinterface.md"
)

add_custom_command(
        OUTPUT
        "${CMAKE_SOURCE_DIR}/docs/password.md"
        "${CMAKE_SOURCE_DIR}/docs/rest.md"
        "${CMAKE_SOURCE_DIR}/docs/log.md"
        "${CMAKE_SOURCE_DIR}/docs/cinterface.md"
        COMMAND go install "github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest"
        COMMAND gomarkdoc "${CMAKE_SOURCE_DIR}" -o "${CMAKE_SOURCE_DIR}/docs/password.md"
        COMMAND gomarkdoc "${CMAKE_SOURCE_DIR}/rest" -o "${CMAKE_SOURCE_DIR}/docs/rest.md"
        COMMAND gomarkdoc "${CMAKE_SOURCE_DIR}/log" -o "${CMAKE_SOURCE_DIR}/docs/log.md"
        COMMAND gomarkdoc "${CMAKE_SOURCE_DIR}/cinterface" -o "${CMAKE_SOURCE_DIR}/docs/cinterface.md"
        DEPENDS ${GO_FILES}
)
