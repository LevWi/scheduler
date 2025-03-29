module;

#include <algorithm>
#include <chrono>
#include <vector>
#include <ranges>

export module IntervalModule;
// import std;

export namespace lf {

template <typename T>
concept isTimePoint =
    requires(T t) { std::chrono::time_point_cast<std::chrono::seconds>(t); };

template <isTimePoint Tp>
struct Interval {
  using TimePoint = Tp;
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

  std::vector<Interval> subtract(
      const Interval &other) const {  // TODO avoid heap allocation?
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

template <isTimePoint Tp>
bool intervals_compare(const Interval<Tp> &a, const Interval<Tp> &b) {
  return a.start < b.start;
};

template <typename Tp>
void sort_by_start(Intervals<Tp> &intervals) {
  std::ranges::sort(intervals, intervals_compare<Tp>);
}

template <typename Tp>
bool is_sorted(const Intervals<Tp> &intervals) {
  return std::ranges::is_sorted(intervals, intervals_compare<Tp>);
}

template <typename Tp>
bool has_overlaps(const Intervals<Tp> &intervals) {
  return std::ranges::adjacent_find(intervals,
                                    [](const auto &a, const auto &b) {
                                      return a.is_overlap(b);
                                    }) != intervals.end();
}

template <typename Tp>
void prepare_united(Intervals<Tp> &intervals) {
  if (intervals.size() < 2) {
    return;
  }
  sort_by_start(intervals);

  auto i = intervals.begin();
  bool united = false;
  for (auto j = i + 1; j < intervals.end(); j++) {
    if (i->is_overlap(*j)) {
      if (i->end < j->end) {
        i->end = j->end;
      }
      united = true;
    } else if (united) {
      i = intervals.erase(i + 1, j);
      j = i;
      united = false;
    } else {
      i = j;
    }
  }
  intervals.erase(i + 1, intervals.end());
}

template<isTimePoint Tp, typename Out = Intervals<Tp>>
Out passed_intervals(std::span<const Interval<Tp>> intervals,
                              std::span<const Interval<Tp>> exclusions) {
  if (intervals.empty()) return {};
  if (exclusions.empty()) return std::ranges::to<Out>(intervals);
  Out out;
  out.reserve(intervals.size());
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

template<std::ranges::random_access_range R>
auto passed_intervals(R&& intervals,
                      R&& exclusions)
{
  using Tp = std::ranges::range_value_t<R>::TimePoint;
  return passed_intervals<Tp>(intervals, exclusions);
}

}  // namespace shed
