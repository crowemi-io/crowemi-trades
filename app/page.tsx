import Feed from "./feed";
import Position from "./position";
import Stat from "./stat";
import Footer from "./footer";
import Header from "./header";

export default function Home() {
  return (
    <div className="mx-auto max-w-7xl sm:px-6 lg:px-8">
      <Header />
      
      <hr className="border-t border-gray-200" />

      <div className="p-14 flex justify-center">
        <Stat />
      </div>

      {/* <hr className="border-t border-gray-200" />

      <div className="pt-14 pb-14">
        <Position />
      </div>

      <hr className="border-t border-gray-200" />

      <div className="pt-20 pb-20 flex justify-center">
        <Feed />
      </div> */}

      <hr className="border-t border-gray-200" />

      <Footer />

    </div>
  );
}
