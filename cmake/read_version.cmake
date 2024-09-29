file(READ "${CMAKE_SOURCE_DIR}/VERSION" TEXT_VERSION)
string(REPLACE "\n" "" TEXT_VERSION "${TEXT_VERSION}")
string(REPLACE "v" "" TEXT_VERSION "${TEXT_VERSION}")
set(
        VERSION
        ${TEXT_VERSION}
        CACHE
        STRING
        "Version number of this project. Usually read from \"VERSION\" file in the root of this project."
)
