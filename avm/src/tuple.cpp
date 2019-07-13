/*
 * Copyright 2019, Offchain Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <avm/tuple.hpp>

#include <avm/util.hpp>

void Tuple::marshal(std::vector<unsigned char>& buf) const {
    buf.push_back(TUPLE + tuple_size());
    for (int i = 0; i < tuple_size(); i++) {
        marshal_value(get_element(i), buf);
    }
}

value Tuple::clone_shallow() {
    Tuple tup(tuplePool, tuple_size());
    for (int i = 0; i < tuple_size(); i++) {
        auto valHash = hash(get_element(i));
        tup.set_element(i, valHash);
    }
    return tup;
}

uint256_t Tuple::calculateHash() const {
    std::array<unsigned char, 1 + 8 * 32> tupData;
    auto oit = tupData.begin();
    tupData[0] = TUPLE + tuple_size();
    ++oit;
    for (int i = 0; i < tuple_size(); i++) {
        auto valHash = hash(get_element(i));
        std::array<uint64_t, 4> valHashInts;
        to_big_endian(valHash, valHashInts.begin());
        std::copy(reinterpret_cast<unsigned char*>(valHashInts.data()),
                  reinterpret_cast<unsigned char*>(valHashInts.data()) + 32,
                  oit);
        oit += 32;
    }

    std::array<unsigned char, 32> hashData;
    evm::Keccak_256(tupData.data(), 1 + 32 * tuple_size(), hashData.data());
    return from_big_endian(hashData.begin(), hashData.end());
}

uint256_t zeroHash() {
    std::array<unsigned char, 1> tupData;
    tupData[0] = TUPLE;
    std::array<unsigned char, 32> hashData;
    evm::Keccak_256(tupData.data(), 1, hashData.data());
    return from_big_endian(hashData.begin(), hashData.end());
}

std::ostream& operator<<(std::ostream& os, const Tuple& val) {
    os << "Tuple(";
    for (int i = 0; i < val.tuple_size(); i++) {
        std::cout << val.get_element(i);
        if (i < val.tuple_size() - 1) {
            os << ", ";
        }
    }
    os << ")";
    return os;
}
