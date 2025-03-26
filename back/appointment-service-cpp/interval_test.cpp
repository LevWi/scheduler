#include <bits/chrono.h>
#include <gtest/gtest.h>

#include <chrono>
import IntervalModule;

using Clock = std::chrono::system_clock;
using TimePoint = Clock::time_point;
using Interval = shed::Interval<TimePoint>;
using Intervals = shed::Intervals<TimePoint>;

using namespace std::chrono;

constinit TimePoint kStartPoint =
    sys_days{2020y / January / 01} + 10h;  // 2020-01-01 10:00:00 // UTC ?

TEST(IntervalTest, BasicOperations) {
  TimePoint start = kStartPoint;
  Interval interval{start, start + 1h};

  EXPECT_TRUE(interval.is_overlap(interval));
  EXPECT_TRUE(interval.is_fit(interval));

  Interval interval2{interval.end, interval.end + 1s};
  EXPECT_FALSE(interval.is_overlap(interval2));
  EXPECT_FALSE(interval.is_fit(interval2));

  interval2.start -= 2s;
  EXPECT_TRUE(interval.is_overlap(interval2));
  EXPECT_FALSE(interval.is_fit(interval2));

  interval2.start = interval.start + 1s;
  interval2.end = interval.end - 1s;
  EXPECT_TRUE(interval.is_overlap(interval2));
  EXPECT_TRUE(interval.is_fit(interval2));
  EXPECT_TRUE(interval2.is_overlap(interval));
  EXPECT_FALSE(interval2.is_fit(interval));

  interval2.start = interval.start + 1s;
  interval2.end = interval.end - 1s;
  EXPECT_TRUE(interval.is_overlap(interval2));
}

TEST(IntervalTest, SubtractTest) {
  TimePoint start = kStartPoint;

  Interval interval{start, start + 1h};
  auto result = interval.subtract(interval);
  EXPECT_TRUE(result.empty());

  // Interval interval2{start + 30min, start + 1h};
  Interval expected{start, start + 30min};
  result = interval.subtract({start + 30min, start + 1h});
  EXPECT_EQ(result.size(), 1);
  EXPECT_EQ(result[0], expected);

  // 76
  result = interval.subtract({start + 31min, start + 1h + 1min});
  expected = Interval{start, start + 31min};
  EXPECT_EQ(result.size(), 1);
  EXPECT_EQ(result[0], expected);

  // 83
  expected = Interval{start + 30min, start + 1h};
  result = interval.subtract({start, start + 30min});
  EXPECT_EQ(result.size(), 1);
  EXPECT_EQ(result[0], expected);
  // 90
  expected = Interval{start + 31min, start + 1h};
  result = interval.subtract({start - 1min, start + 31min});
  EXPECT_EQ(result.size(), 1);
  EXPECT_EQ(result[0], expected);
  // 97
  auto expected2 = Intervals{
      {start, start + 10min},
      {start + 1h - 10min, start + 1h},
  };
  result = interval.subtract({start + 10min, start + 1h - 10min});
  EXPECT_EQ(result.size(), 2);
  EXPECT_EQ(result[0], expected2[0]);
  EXPECT_EQ(result[1], expected2[1]);
}

int main(int argc, char **argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
