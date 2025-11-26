FROM alpine

WORKDIR /app

COPY paymentServiceApp .

RUN chmod +x paymentServiceApp

CMD ["./paymentServiceApp"]