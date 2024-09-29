if (NOT DEFINED INSTALL_INCLUDEDIR)
    set(INSTALL_INCLUDEDIR "include")
endif ()
if (NOT INSTALL_INCLUDEDIR MATCHES "/$")
    set(INSTALL_INCLUDEDIR "${INSTALL_INCLUDEDIR}/")
endif ()


include(GNUInstallDirs)
include(CMakePackageConfigHelpers)

write_basic_package_version_file(
        "${CMAKE_BINARY_DIR}/${CMAKE_PROJECT_NAME}ConfigVersion.cmake"
        VERSION ${CMAKE_PROJECT_VERSION}
        COMPATIBILITY SameMajorVersion
)

configure_package_config_file(
        "${CMAKE_SOURCE_DIR}/cmake/installConfig.cmake.in"
        "${CMAKE_BINARY_DIR}/${CMAKE_PROJECT_NAME}Config.cmake"
        INSTALL_DESTINATION "${CMAKE_INSTALL_LIBDIR}/cmake/${CMAKE_PROJECT_NAME}-${CMAKE_PROJECT_VERSION}"
)

install(
        FILES
        "${DLL_FILEPATH}"
        DESTINATION "${CMAKE_INSTALL_LIBDIR}/${CMAKE_PROJECT_NAME}-${CMAKE_PROJECT_VERSION}"
)

if (WIN32)
    install(
            FILES
            "${LIB_FILEPATH}"
            DESTINATION "${CMAKE_INSTALL_LIBDIR}/${CMAKE_PROJECT_NAME}-${CMAKE_PROJECT_VERSION}"
    )
endif ()

install(
        FILES
        "${CMAKE_BINARY_DIR}/${CMAKE_PROJECT_NAME}ConfigVersion.cmake"
        "${CMAKE_BINARY_DIR}/${CMAKE_PROJECT_NAME}Config.cmake"
        DESTINATION "${CMAKE_INSTALL_LIBDIR}/cmake/${CMAKE_PROJECT_NAME}-${CMAKE_PROJECT_VERSION}"
)

install(
        DIRECTORY "${INSTALL_INCLUDEDIR}"
        DESTINATION "${CMAKE_INSTALL_INCLUDEDIR}/${CMAKE_PROJECT_NAME}-${CMAKE_PROJECT_VERSION}"
        FILES_MATCHING PATTERN "*.h"
)

install(
        DIRECTORY "${INSTALL_INCLUDEDIR}"
        DESTINATION "${CMAKE_INSTALL_INCLUDEDIR}/${CMAKE_PROJECT_NAME}-${CMAKE_PROJECT_VERSION}"
        FILES_MATCHING PATTERN "*.hpp"
)
