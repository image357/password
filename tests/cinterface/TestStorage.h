#ifndef TESTSTORAGE_H
#define TESTSTORAGE_H

#include <gtest/gtest.h>
#include <gmock/gmock.h>

class TestStorage : public ::testing::Test {
protected:
    void SetUp() override;

    void TearDown() override;
};



#endif //TESTSTORAGE_H
