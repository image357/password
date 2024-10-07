cmake_minimum_required(VERSION 3.20)
include(cmake/read_version.cmake)

# options

# start project
project(password VERSION ${VERSION})
set(CMAKE_CXX_STANDARD 17)

if (NOT WIN32 AND NOT UNIX)
    message(FATAL_ERROR "unknown OS")
endif ()

if (WIN32 AND NOT MINGW)
    message(FATAL_ERROR "go build needs MinGW")
endif ()

# build variables
set(
        GO_FILES
        password.go
        storage.go
        encryption.go
        rest/simple.go
        rest/multi.go
        rest/access.go
        log/log.go
        cinterface/cinterface.go
)

set(LIB_NAME "cinterface")

if (WIN32)
    set(DLL_FILE "lib${LIB_NAME}.dll")
endif ()
if (UNIX)
    set(DLL_FILE "lib${LIB_NAME}.so")
endif ()
set(DLL_FILEPATH "${CMAKE_BINARY_DIR}/${DLL_FILE}")

set(GIN_FILE "lib${LIB_NAME}.h")
set(GIN_FILEPATH "${CMAKE_BINARY_DIR}/${GIN_FILE}")

set(INC_FILE "${LIB_NAME}.h")
set(INC_DIR "${CMAKE_BINARY_DIR}/include/${CMAKE_PROJECT_NAME}")
set(INC_FILEPATH "${INC_DIR}/${INC_FILE}")

set(TEST_BINDIR "${CMAKE_BINARY_DIR}/tests/cinterface")
set(DLL_TESTPATH "${TEST_BINDIR}/${DLL_FILE}")

if (WIN32)
    set(DEF_FILE "lib${LIB_NAME}.def")
    set(DEF_FILEPATH "${CMAKE_BINARY_DIR}/${DEF_FILE}")

    set(LIB_FILE "lib${LIB_NAME}.lib")
    set(LIB_FILEPATH "${CMAKE_BINARY_DIR}/${LIB_FILE}")
endif ()

# custom go build target
add_custom_target(
        go_build_cinterface ALL
        DEPENDS
        "${DLL_FILEPATH}"
        "${INC_FILEPATH}"
        "${DLL_TESTPATH}"
        "${LIB_FILEPATH}"
)

add_custom_command(
        OUTPUT
        "${DLL_FILEPATH}"
        "${GIN_FILEPATH}"
        "${INC_FILEPATH}"
        "${DLL_TESTPATH}"
        COMMAND go mod tidy
        COMMAND go build -buildmode=c-shared -o "${DLL_FILEPATH}" "github.com/image357/password/cinterface"
        COMMAND go build -o "${CMAKE_BINARY_DIR}/" "github.com/image357/password/cmd/patchheader"
        COMMAND "${CMAKE_BINARY_DIR}/patchheader" "${GIN_FILEPATH}"
        COMMAND ${CMAKE_COMMAND} -E make_directory "${INC_DIR}"
        COMMAND ${CMAKE_COMMAND} -E copy_if_different "${GIN_FILEPATH}" "${INC_FILEPATH}"
        COMMAND ${CMAKE_COMMAND} -E make_directory "${TEST_BINDIR}"
        COMMAND ${CMAKE_COMMAND} -E copy_if_different "${DLL_FILEPATH}" "${DLL_TESTPATH}"
        DEPENDS ${GO_FILES}
)

if (WIN32)
    add_custom_command(
            OUTPUT
            "${DEF_FILEPATH}"
            "${LIB_FILEPATH}"
            COMMAND gendef "${DLL_FILEPATH}"
            COMMAND dlltool -k --output-lib "${LIB_FILEPATH}" --def "${DEF_FILEPATH}"
            DEPENDS "${DLL_FILEPATH}"
    )
endif ()

# prepare include directory
file(MAKE_DIRECTORY "${INC_DIR}")

# create cmake library from custom target
add_library(${LIB_NAME} SHARED IMPORTED)
set_property(
        TARGET ${LIB_NAME} PROPERTY
        IMPORTED_LOCATION "${DLL_FILEPATH}"
)

if (WIN32)
    set_property(
            TARGET ${LIB_NAME} PROPERTY
            IMPORTED_IMPLIB "${LIB_FILEPATH}"
    )
endif ()

include(GNUInstallDirs)
target_include_directories(
        ${LIB_NAME} INTERFACE
        $<BUILD_INTERFACE:${CMAKE_BINARY_DIR}/include>
        $<INSTALL_INTERFACE:${CMAKE_INSTALL_INCLUDEDIR}/${CMAKE_PROJECT_NAME}-${CMAKE_PROJECT_VERSION}>
)

add_dependencies(${LIB_NAME} go_build_cinterface)

add_subdirectory(tests/cinterface EXCLUDE_FROM_ALL)


# install routines
set(INSTALL_INCLUDEDIR "${CMAKE_BINARY_DIR}/include")
set(INSTALL_LIB_TARGET ${LIB_NAME})
include(cmake/install_targets.cmake)
include(cmake/install_license.cmake)
