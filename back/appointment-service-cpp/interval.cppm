module;

#include <algorithm>
#include <chrono>
#include <vector>

export module IntervalModule;
// import std;

export namespace shed {

template <typename T>
concept isTimePoint =
    requires(T t) { std::chrono::time_point_cast<std::chrono::seconds>(t); };

template <isTimePoint Tp>
struct Interval {
  Tp start{};
  Tp end{};

  bool is_valid() const { return start < end; }

  bool is_empty() const { return start == Tp{} && end == Tp{}; }

  bool is_overlap(const Interval &other) const {
    return start < other.end && end > other.start;
  }

  bool before(const Interval &other) const { return end <= other.start; }

  bool is_fit(const Interval &other) const {
    return start <= other.start && end >= other.end;
  }

  std::vector<Interval> subtract(const Interval &other) const {
    if (!is_overlap(other)) return {*this};
    if (other.is_fit(*this)) return {};

    std::vector<Interval> result;
    if (start < other.start) {
      if (end <= other.end) {
        result.push_back({start, other.start});
      } else {
        result.push_back({start, other.start});
        result.push_back({other.end, end});
      }
    } else {
      result.push_back({other.end, end});
    }
    return result;
  }

  Interval intersection(const Interval &other) const {
    return {std::max(start, other.start), std::min(end, other.end)};
  }

  bool operator==(const Interval &) const = default;
};

template <isTimePoint Tp>
using Intervals = std::vector<Interval<Tp>>;

template <typename Tp>
void sort_by_start(Intervals<Tp> &intervals) {
  std::ranges::sort(intervals, [](const auto &a, const auto &b) {
    return a.start < b.start;
  });
}

template <typename Tp>
bool has_overlaps(const Intervals<Tp> &intervals) {
  for (std::size_t i = 0; i < intervals.size() - 1; ++i) {
    if (intervals[i].is_overlap(intervals[i + 1])) {
      return true;
    }
  }
  return false;
}

template <typename Tp>
Intervals<Tp> prepare_united(Intervals<Tp> intervals) {
  sort_by_start(intervals);
  Intervals<Tp> result;
  for (const auto &interval : intervals) {
    if (!result.empty() && result.back().is_overlap(interval)) {
      result.back().end = std::max(result.back().end, interval.end);
    } else {
      result.push_back(interval);
    }
  }
  return result;
}

template <typename Tp>
Intervals<Tp> passed_intervals(const Intervals<Tp> &intervals,
                               const Intervals<Tp> &exclusions) {
  if (intervals.empty()) return {};
  if (exclusions.empty()) return intervals;

  Intervals<Tp> out;
  std::size_t exclusionIndex = 0;
  Interval<Tp> tmp{};

  for (std::size_t i = 0; i < intervals.size() || tmp.is_valid();) {
    if (!tmp.is_valid() && i < intervals.size()) {
      tmp = intervals[i++];
    }
    if (exclusionIndex >= exclusions.size() ||
        tmp.before(exclusions[exclusionIndex])) {
      out.push_back(tmp);
      tmp = {};
    } else if (exclusions[exclusionIndex].before(tmp)) {
      ++exclusionIndex;
    } else if (tmp.is_overlap(exclusions[exclusionIndex])) {
      auto result = tmp.subtract(exclusions[exclusionIndex]);
      if (result.size() == 2) {
        out.push_back(result[0]);
        tmp = result[1];
      } else if (result.size() == 1) {
        if (result[0].before(exclusions[exclusionIndex])) {
          out.push_back(result[0]);
          tmp = {};
        } else {
          tmp = result[0];
        }
      } else {
        tmp = {};
      }
    } else {
      throw std::logic_error("Unexpected overlap condition");
    }
  }

  if (tmp.is_valid()) {
    out.push_back(tmp);
  }

  return out;
}
}  // namespace shed
