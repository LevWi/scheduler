#include <bits/chrono.h>
#include <gtest/gtest.h>

#include <chrono>
#include <vector>
import IntervalModule;

using Clock = std::chrono::system_clock;
using TimePoint = Clock::time_point;
using Interval = ::lf::Interval<TimePoint>;
using Intervals = ::lf::Intervals<TimePoint>;

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

TimePoint make_time(TimePoint::duration seconds) {
  return kStartPoint + seconds;
}

TEST(PrepareUnitedTests, MergesOverlappingIntervals) {
  Intervals intervals = {{make_time(0min), make_time(10min)},
                         {make_time(5min), make_time(15min)},
                         {make_time(12min), make_time(20min)},
                         {make_time(18min), make_time(25min)},
                         {make_time(30min), make_time(40min)},
                         {make_time(35min), make_time(50min)}};

  prepare_united(intervals);

  ASSERT_EQ(intervals.size(), 2);
  EXPECT_EQ(intervals[0].start, make_time(0min));
  EXPECT_EQ(intervals[0].end, make_time(25min));
  EXPECT_EQ(intervals[1].start, make_time(30min));
  EXPECT_EQ(intervals[1].end, make_time(50min));

  intervals = {{make_time(0min), make_time(10min)},
               {make_time(5min), make_time(15min)},
               {make_time(15min), make_time(20min)}};

  prepare_united(intervals);

  ASSERT_EQ(intervals.size(), 2);
  EXPECT_EQ(intervals[0].start, make_time(0min));
  EXPECT_EQ(intervals[0].end, make_time(15min));
  EXPECT_EQ(intervals[1].start, make_time(15min));
  EXPECT_EQ(intervals[1].end, make_time(20min));
}

TEST(PrepareUnitedTests, KeepsNonOverlappingIntervals) {
  Intervals intervals = {{make_time(0s), make_time(5s)},
                         {make_time(10s), make_time(15s)},
                         {make_time(20s), make_time(25s)},
                         {make_time(30s), make_time(35s)},
                         {make_time(40s), make_time(45s)}};

  auto copy = intervals;
  prepare_united(intervals);

  ASSERT_EQ(intervals.size(), 5);  // No merging should occur
  ASSERT_TRUE(std::ranges::equal(intervals, copy));

  intervals = {
      {make_time(0s), make_time(10s)},
      {make_time(10s), make_time(20s)},
      {make_time(20s), make_time(30s)},
      {make_time(30s), make_time(40s)},
  };

  copy = intervals;
  prepare_united(intervals);

  ASSERT_EQ(intervals.size(), 4);
  ASSERT_TRUE(std::ranges::equal(intervals, copy));
}

TEST(PrepareUnitedTests, RemainsUnchanged) {
  Intervals intervals;
  prepare_united(intervals);
  ASSERT_TRUE(intervals.empty());

  intervals = {{make_time(0s), make_time(10s)}};

  prepare_united(intervals);

  ASSERT_EQ(intervals.size(), 1);
  EXPECT_EQ(intervals[0].start, make_time(0s));
  EXPECT_EQ(intervals[0].end, make_time(10s));
}

TEST(PrepareUnitedTests, MixedMergingAndNonMerging) {
  Intervals intervals = {
      {make_time(0s), make_time(10s)},
      {make_time(5s), make_time(12s)},  // Overlaps with first
      {make_time(20s), make_time(25s)},
      {make_time(30s), make_time(35s)},
      {make_time(32s), make_time(40s)},  // Overlaps with previous
      {make_time(40s), make_time(60s)}};

  prepare_united(intervals);

  ASSERT_EQ(intervals.size(), 4);
  EXPECT_EQ(intervals[0].start, make_time(0s));
  EXPECT_EQ(intervals[0].end, make_time(12s));
  EXPECT_EQ(intervals[1].start, make_time(20s));
  EXPECT_EQ(intervals[1].end, make_time(25s));
  EXPECT_EQ(intervals[2].start, make_time(30s));
  EXPECT_EQ(intervals[2].end, make_time(40s));
  EXPECT_EQ(intervals[3].start, make_time(40s));
  EXPECT_EQ(intervals[3].end, make_time(60s));
}

TEST(IntervalsTest, TestSetPassedIntervals) {
  auto checkCase = [](const Intervals& i, const Intervals& e,
                      const Intervals& expected) {
    auto result = passed_intervals(i, e);
    EXPECT_TRUE(std::ranges::equal(result, expected))
        << "Expected: " << expected.size()
        << " intervals, got: " << result.size();
  };

  const auto start_time = sys_days{2024y / October / 9};
  // 157
  checkCase({{{start_time + 9h}, {start_time + 18h}}},
            {{{start_time + 12h}, {start_time + 13h}}},
            {{{start_time + 9h}, {start_time + 12h}},
             {{start_time + 13h}, {start_time + 18h}}});
  // 181
  checkCase({{{start_time + 9h}, {start_time + 18h}}},
            {{{start_time + 12h}, {start_time + 18h}}},
            {{{start_time + 9h}, {start_time + 12h}}});

  // 203
  checkCase(
      {
          {
              {start_time + 9h},
              {start_time + 18h},
          },
      },
      {
          {
              {start_time + 8h},
              {start_time + 18h},
          },
      },
      Intervals{});

  // 220
  checkCase({},
            {
                {
                    {start_time + 8h},
                    {start_time + 18h},
                },
            },
            {});

  // 232
  checkCase(
      {
          {
              start_time + 9h,
              start_time + 18h,
          },
          {
              start_time + 20h,
              start_time + 21h,
          },
      },
      {
          {
              start_time + 10h,
              start_time + 11h,
          },
          {
              start_time + 13h,
              start_time + 14h,
          },
          {
              start_time + 15h,
              start_time + 16h,
          },
      },
      {
          {
              start_time + 9h,
              start_time + 10h,
          },
          {
              start_time + 11h,
              start_time + 13h,
          },
          {
              start_time + 14h,
              start_time + 15h,
          },
          {
              start_time + 16h,
              start_time + 18h,
          },
          {
              start_time + 20h,
              start_time + 21h,
          },
      });
}

int main(int argc, char** argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
