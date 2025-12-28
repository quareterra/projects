#include <cmath>
#include <fstream>
#include <iostream>

struct complex_t {
  double r{}, i{};
  complex_t(double r, double i) : r{r}, i{i} {}

  complex_t pow2() {
    return complex_t{std::pow(r, 2) - std::pow(i, 2), 2 * r * i};
  }
};

std::ostream &operator<<(std::ostream &stream, const complex_t& complex) {
  stream << '(' << complex.r << " + " << complex.i << "i)";
  return stream;
}

complex_t operator+(const complex_t &left, const complex_t &right) {
  return complex_t{left.r + right.r, left.i + right.i};
}

int draw(double u, double v) {
  complex_t z{0, 0}, c{u, v};
  int i = 0;
  for (; i < 4000; ++i) {
    z = z.pow2() + c;
    if ((z.r == 0 && z.i == 0) || z.r > 1e10 || z.i > 1e10) break;
  }
  return i;
}

int main(int, char **) {
  std::ofstream image("image.ppm", std::ios::binary);
  const auto size = 2048;
  const auto width = size, height = size;
  image << "P6\n" << width << ' ' << height << '\n' << 255 << '\n';

  for (size_t y = 0; y < height; ++y) {
    for (size_t x = 0; x < width; ++x) {
      double u = (static_cast<double>(x) / static_cast<double>(width) - 0.5) *
                 4.0,
             v = (static_cast<double>(y) / static_cast<double>(height) - 0.5) *
                 4.0;
      unsigned char r{}, g{}, b{};
      auto number = draw(u, v);
      r = static_cast<unsigned char>((number % 3) == 0) * 255;
      g = static_cast<unsigned char>((number % 3) == 1) * 255;
      b = static_cast<unsigned char>((number % 3) == 2) * 255;
      image << r << g << b;
    }
  }

  image.close();

  return EXIT_SUCCESS;
}
