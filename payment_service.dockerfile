FROM alpine

WORKDIR /app

COPY paymentServiceApp .

CMD ["./paymentServiceApp"]