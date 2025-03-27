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

TEST(IntervalTest, Subtract) {
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

TEST(IntervalTest, SortingAndOverlap) {
  Interval interval{kStartPoint, kStartPoint + 1min};

  Intervals intervals = {interval, interval};
  sort_by_start(intervals);
  EXPECT_TRUE(is_sorted(intervals));
  EXPECT_TRUE(has_overlaps(intervals));

  intervals.clear();
  for (int i = 0; i < 10; ++i) {
    auto tp = kStartPoint + i * 1min;
    intervals.push_back({tp, tp + 1min});
  }
  sort_by_start(intervals);
  EXPECT_TRUE(is_sorted(intervals));
  EXPECT_FALSE(has_overlaps(intervals));

  intervals.push_back(intervals[0]);
  EXPECT_FALSE(is_sorted(intervals));
  sort_by_start(intervals);
  EXPECT_TRUE(is_sorted(intervals));
  EXPECT_TRUE(has_overlaps(intervals));
}

TEST(IntervalsTest, TestSetPassedIntervals) {
  auto checkCase = [](const Intervals& i, const Intervals& e,
                      const Intervals& expected) {
    auto result = passed_intervals(i, e);
    EXPECT_TRUE(std::ranges::equal(result, expected))
        << "Expected: " << expected.size()
        << " intervals, got: " << result.size();
  };

  // 157
  checkCase({{{sys_days{2024y / October / 9} + 9h},
              {sys_days{2024y / October / 9} + 18h}}},
            {{{sys_days{2024y / October / 9} + 12h},
              {sys_days{2024y / October / 9} + 13h}}},
            {{{sys_days{2024y / October / 9} + 9h},
              {sys_days{2024y / October / 9} + 12h}},
             {{sys_days{2024y / October / 9} + 13h},
              {sys_days{2024y / October / 9} + 18h}}});
  // 181
  checkCase({{{sys_days{2024y / October / 9} + 9h},
              {sys_days{2024y / October / 9} + 18h}}},
            {{{sys_days{2024y / October / 9} + 12h},
              {sys_days{2024y / October / 9} + 18h}}},
            {{{sys_days{2024y / October / 9} + 9h},
              {sys_days{2024y / October / 9} + 12h}}});

  // 203
  checkCase(
      {
          {
              {sys_days{2024y / October / 9} + 9h},
              {sys_days{2024y / October / 9} + 18h},
          },
      },
      {
          {
              {sys_days{2024y / October / 9} + 8h},
              {sys_days{2024y / October / 9} + 18h},
          },
      },
      Intervals{});

  // 220
  checkCase({},
            {
                {
                    {sys_days{2024y / October / 9} + 8h},
                    {sys_days{2024y / October / 9} + 18h},
                },
            },
            {});

  // 232
  checkCase(
      {
          {
              sys_days{2024y / October / 9} + 9h,
              sys_days{2024y / October / 9} + 18h,
          },
          {
              sys_days{2024y / October / 9} + 20h,
              sys_days{2024y / October / 9} + 21h,
          },
      },
      {
          {
              sys_days{2024y / October / 9} + 10h,
              sys_days{2024y / October / 9} + 11h,
          },
          {
              sys_days{2024y / October / 9} + 13h,
              sys_days{2024y / October / 9} + 14h,
          },
          {
              sys_days{2024y / October / 9} + 15h,
              sys_days{2024y / October / 9} + 16h,
          },
      },
      {
          {
              sys_days{2024y / October / 9} + 9h,
              sys_days{2024y / October / 9} + 10h,
          },
          {
              sys_days{2024y / October / 9} + 11h,
              sys_days{2024y / October / 9} + 13h,
          },
          {
              sys_days{2024y / October / 9} + 14h,
              sys_days{2024y / October / 9} + 15h,
          },
          {
              sys_days{2024y / October / 9} + 16h,
              sys_days{2024y / October / 9} + 18h,
          },
          {
              sys_days{2024y / October / 9} + 20h,
              sys_days{2024y / October / 9} + 21h,
          },
      });
}

int main(int argc, char** argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
