#ifndef TESTREST_H
#define TESTREST_H

#include <gtest/gtest.h>
#include <gmock/gmock.h>

class TestRest : public ::testing::Test {
protected:
    void SetUp() override;

    void TearDown() override;
};

bool test_callback(const char *token, const char *ip, const char *resource, const char *id);

#endif //TESTREST_H
