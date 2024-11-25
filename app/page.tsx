import Feed from "./feed";
import Position from "./position";
import Stat from "./stat";
import Footer from "./footer";

export default function Home() {
  return (
    <div className="mx-auto max-w-7xl sm:px-6 lg:px-8">
      <div className="p-20 flex justify-center">
        <Stat />
      </div>

      <hr className="border-t border-gray-200" />

      <div className="pt-20 pb-20">
        <Position />
      </div>

      <hr className="border-t border-gray-200" />

      <div className="pt-20 pb-20 flex justify-center">
        <Feed />
      </div>
      <Footer></Footer>

    </div>
  );
}
