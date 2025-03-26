cmake_minimum_required(VERSION 3.28)

project(appointment_service
        VERSION 0.1
        LANGUAGES CXX)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

include(GNUInstallDirs)

add_library(base INTERFACE IMPORTED)
add_library(lf::base ALIAS base)
target_compile_features(base INTERFACE cxx_std_20)


add_library(core OBJECT)
target_link_libraries(core PUBLIC lf::base)
target_sources(core PUBLIC
    FILE_SET core_module
    TYPE CXX_MODULES
    FILES
        interval.cppm
)


include(CTest)
if(BUILD_TESTING)
    add_executable(service_test
        #$<TARGET_OBJECTS:core>
        interval_test.cpp
    )
    target_link_libraries(service_test PUBLIC gtest core)

    add_test(NAME Test COMMAND service_test)
endif()
