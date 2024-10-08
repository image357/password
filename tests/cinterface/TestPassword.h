#ifndef TEST_PASSWORD_H
#define TEST_PASSWORD_H

#include <gtest/gtest.h>
#include <gmock/gmock.h>

class TestPassword : public ::testing::Test {
protected:
    void SetUp() override;

    void TearDown() override;
};

#endif //TEST_PASSWORD_H
